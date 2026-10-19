#include "library.hpp"
#include <fstream>
#include <iostream>
#include <stdexcept>
#include <unistd.h>

using namespace picker;
void require(bool value, const char* message) {
    if (!value) throw std::runtime_error(message);
}
void write(const fs::path& path, const std::string& content) {
    fs::create_directories(path.parent_path());
    std::ofstream(path) << content;
}
int main() {
    char pattern[] = "/tmp/picker-test-XXXXXX";
    const fs::path root = mkdtemp(pattern);
    try {
        Locations locations{root / "user", root / "shared", root / "legacy", root};
        fs::create_directories(locations.shared);
        const auto roms = root / "not-a-standard-rom-root";
        write(roms / "odd 'name; $.rom", "fixture");
        write(roms / "second.rom", "fixture");
        write(roms / "hidden.rom", "fixture");
        write(locations.shared / "es_systems.cfg", "<systemList><system><name>wrong</name><path>/wrong</path></system></systemList>");
        write(locations.user / "es_systems_custom.cfg", "<systemList><system><name>custom</name><fullname>Custom Console</fullname><path>" + roms.string() + "</path></system></systemList>");
        write(roms / "gamelist.xml", "<gameList><game><path>./odd 'name; $.rom</path><name>Scraped Name</name><image>./images/box.png</image><favorite>true</favorite></game></gameList>");
        write(roms / "images/box.png", "fixture art");
        Json systems = Json::array({{{"name", "custom"}, {"gamesystem", "true"}, {"collection", "false"}},
                                    {{"name", "favorites"}, {"gamesystem", "true"}, {"collection", "true"}}});
        Json games = Json::array({{{"path", (roms / "odd 'name; $.rom").string()}, {"name", "Live Title"}, {"favorite", "true"}},
                                  {{"path", (roms / "second.rom").string()}, {"name", "Second"}},
                                  {{"path", (roms / "hidden.rom").string()}, {"hidden", "true"}},
                                  {{"path", (roms / "deleted.rom").string()}}});
        auto fetch = [&](const std::string& route) {
            if (route == "/systems") return systems.dump();
            if (route == "/systems/custom/games") return games.dump();
            throw std::runtime_error("unexpected route: " + route);
        };
        auto library = discover(locations, fetch);
        require(library.games.size() == 2, "custom roots, hidden/deleted paths and collection filtering");
        require(library.games[0].title == "Live Title", "live ES metadata takes priority");
        require(library.games[0].image == (roms / "images/box.png").string(), "art resolves against ROM root");
        require(eligible(library, "custom", true, {}).size() == 1, "favorite string flag");
        require(eligible(library, "custom", true, {library.games[0].path}).empty(), "recent exclusion intersects favorites");
        require(eligible(library, "missing", false, {}).empty(), "unknown filter is empty");
        require(encode_segment("a b/#") == "a%20b%2F%23", "API segment escaping");
        systems.push_back({{"name", "alias"}, {"gamesystem", "true"}, {"collection", "false"}});
        write(locations.user / "es_systems_alias.cfg", "<systemList><system><name>alias</name><path>" + roms.string() + "</path></system></systemList>");
        auto alias_fetch = [&](const std::string& route) {
            if (route == "/systems/alias/games") return Json::array({games[0]}).dump();
            return fetch(route);
        };
        library = discover(locations, alias_fetch);
        require(library.games.size() == 1 && library.games[0].title == "Second", "ambiguous launch paths must be blocked across systems");

        Library uneven;
        for (int i = 0; i < 100; ++i)
            uneven.games.push_back({i == 0 ? "small" : "large", "", std::to_string(i), "", "", false});
        std::mt19937 rng(9123);
        const auto pool = eligible(uneven, "", false, {});
        int small = 0;
        for (int i = 0; i < 20000; ++i) small += choose(pool, rng) == 0;
        require(small > 100 && small < 300, "flat title sampling, not system-first sampling");
        require(choose({72}, rng) == 72, "single-entry pool preserves library index");
        bool empty_rejected = false;
        try { choose({}, rng); } catch (const std::exception&) { empty_rejected = true; }
        require(empty_rejected, "empty pool is explicit");
        fs::remove_all(root);
        std::cout << "PASS: config/gamelist/live catalogue, path ambiguity, filters, uniform selection\n";
        return 0;
    } catch (const std::exception& error) {
        fs::remove_all(root);
        std::cerr << error.what() << '\n';
        return 1;
    }
}
