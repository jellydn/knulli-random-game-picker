#include "library.hpp"
#include <algorithm>
#include <map>
#include <pugixml.hpp>
#include <sstream>
#include <stdexcept>
#include <unordered_map>
#include <unordered_set>

namespace picker {
namespace {
bool flag(const Json& value, const char* key) {
    auto found = value.find(key);
    return found != value.end() && (*found == true || *found == "true");
}
std::string string(const Json& value, const char* key) {
    const auto found = value.find(key);
    return found != value.end() && found->is_string() ? found->get<std::string>() : "";
}
fs::path resolve(std::string text, const fs::path& root, const fs::path& home) {
    if (text.rfind("~/", 0) == 0) return (home / text.substr(2)).lexically_normal();
    const fs::path path(text);
    return (path.is_absolute() ? path : root / path).lexically_normal();
}
std::string identity(const fs::path& path) {
    std::error_code error;
    auto canonical = fs::weakly_canonical(path, error);
    return (error ? path.lexically_normal() : canonical).string();
}
bool path_exists(const fs::path& path) {
    std::error_code error;
    return fs::exists(path, error);
}
void parse_xml(pugi::xml_document& doc, const fs::path& path) {
    std::error_code error;
    if (fs::file_size(path, error) > 32 * 1024 * 1024 || error)
        throw std::runtime_error("Cannot read XML (32 MiB limit): " + path.string());
    if (!doc.load_file(path.c_str())) throw std::runtime_error("Invalid XML: " + path.string());
}
}

Library discover(const Locations& locations, const Fetch& fetch) {
    Library library;
    const auto own_launcher = identity(fs::canonical("/proc/self/exe").parent_path() / "RandomGamePicker.sh");
    pugi::xml_document config;
    fs::path base;
    for (const auto& path : {locations.user / "es_systems_custom.cfg", locations.user / "es_systems.cfg",
                             locations.shared / "es_systems.cfg", locations.legacy / "es_systems.cfg"}) {
        if (path_exists(path)) { base = path; break; }
    }
    if (base.empty()) throw std::runtime_error("No EmulationStation system configuration found.");
    parse_xml(config, base);
    auto systems = config.child("systemList");
    if (!systems) throw std::runtime_error("System configuration has no systemList.");
    for (const auto& directory : {locations.user, locations.shared}) {
        std::error_code error;
        for (const auto& entry : fs::directory_iterator(directory, error)) {
            const auto path = entry.path();
            if (path.extension() != ".cfg" || path.filename().string().rfind("es_systems_", 0) != 0) continue;
            pugi::xml_document extra;
            parse_xml(extra, path);
            for (const auto& addition : extra.child("systemList").children("system")) {
                const std::string name = addition.child_value("name");
                if (name.empty()) continue;
                pugi::xml_node target;
                for (auto node : systems.children("system"))
                    if (name == node.child_value("name")) { target = node; break; }
                if (!target) { systems.append_copy(addition); continue; }
                for (auto field : addition.children()) {
                    if (std::string(field.name()) == "name") continue;
                    target.remove_child(field.name());
                    if (std::string(field.name()) == "emulators" || *field.child_value()) target.append_copy(field);
                }
            }
        }
    }
    const Json live_systems = Json::parse(fetch("/systems"));
    if (!live_systems.is_array()) throw std::runtime_error("ES /systems returned an unsupported response.");
    std::unordered_map<std::string, int> path_counts;
    for (const auto& live_system : live_systems) {
        if (flag(live_system, "collection") || !flag(live_system, "gamesystem")) continue;
        const std::string id = string(live_system, "name");
        if (id.empty()) continue;
        const Json live_games = Json::parse(fetch("/systems/" + encode_segment(id) + "/games"));
        if (!live_games.is_array()) throw std::runtime_error("ES returned an unsupported games response.");
        pugi::xml_node system;
        for (auto node : systems.children("system"))
            if (id == node.child_value("name")) { system = node; break; }
        const fs::path root = resolve(system.child_value("path"), locations.home, locations.home);
        const std::string label = *system.child_value("fullname") ? system.child_value("fullname") : id;
        const bool supported = system && *system.child_value("path") && root.is_absolute() && id != "ports";
        pugi::xml_document gamelist;
        std::unordered_map<std::string, pugi::xml_node> metadata;
        if (supported) {
            auto path = root / "gamelist.xml";
            if (!path_exists(path)) path = locations.user / "gamelists" / id / "gamelist.xml";
            if (path_exists(path)) {
                try {
                    parse_xml(gamelist, path);
                    for (auto node : gamelist.child("gameList").children("game"))
                        metadata[identity(resolve(node.child_value("path"), root, locations.home))] = node;
                } catch (const std::exception& error) { library.warnings.push_back(error.what()); }
            }
        }
        for (const auto& live_game : live_games) {
            const auto path = string(live_game, "path");
            if (path.empty() || path.size() > 4096 || path.find('\0') != std::string::npos) continue;
            // The launch API searches every real system, even unsupported/hidden games.
            const auto key = identity(path);
            ++path_counts[key];
            if (!supported || key == own_launcher || !fs::path(path).is_absolute() || !path_exists(path) || flag(live_game, "hidden")) continue;
            auto node = metadata[key];
            Game game{id, label, path, string(live_game, "name"), "", flag(live_game, "favorite")};
            if (game.title.empty()) game.title = *node.child_value("name") ? node.child_value("name") : fs::path(path).stem().string();
            if (game.title.size() > 4096) game.title.resize(4096);
            for (const char* tag : {"image", "thumbnail", "marquee"}) {
                if (!*node.child_value(tag)) continue;
                const auto art = resolve(node.child_value(tag), root, locations.home);
                if (path_exists(art)) { game.image = art.string(); break; }
            }
            library.games.push_back(std::move(game));
        }
    }
    const auto before = library.games.size();
    library.games.erase(std::remove_if(library.games.begin(), library.games.end(), [&](const Game& game) {
        return path_counts[identity(game.path)] != 1;
    }), library.games.end());
    if (before != library.games.size()) library.warnings.push_back("Ambiguous paths excluded to prevent multiple launches.");
    return library;
}

std::vector<std::size_t> eligible(const Library& library, const std::string& system, bool favorites,
                                const std::vector<std::string>& recent) {
    const std::unordered_set<std::string> excluded(recent.begin(), recent.end());
    std::vector<std::size_t> result;
    for (std::size_t i = 0; i < library.games.size(); ++i) {
        const auto& game = library.games[i];
        if ((!system.empty() && game.system != system) || (favorites && !game.favorite) || excluded.count(game.path)) continue;
        result.push_back(i);
    }
    return result;
}
std::size_t choose(const std::vector<std::size_t>& pool, std::mt19937& engine) {
    if (pool.empty()) throw std::runtime_error("No eligible games. Change filters or reset recent selections.");
    return pool[std::uniform_int_distribution<std::size_t>(0, pool.size() - 1)(engine)];
}
std::string encode_segment(const std::string& text) {
    constexpr char hex[] = "0123456789ABCDEF";
    std::string result;
    for (const unsigned char ch : text) {
        if ((ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z') || (ch >= '0' && ch <= '9') || ch == '-' || ch == '_' || ch == '.') result += ch;
        else { result += '%'; result += hex[ch >> 4]; result += hex[ch & 15]; }
    }
    return result;
}
}
