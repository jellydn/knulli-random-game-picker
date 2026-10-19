#include "http.hpp"
#include "library.hpp"
#include <SDL.h>
#include <SDL_image.h>
#include <SDL_ttf.h>
#include <algorithm>
#include <atomic>
#include <fstream>
#include <future>
#include <iostream>
#include <memory>
#include <optional>
#include <set>
#include <thread>

using namespace picker;
namespace {
constexpr SDL_Color white{229, 239, 242, 255}, muted{156, 178, 186, 255}, accent{70, 220, 174, 255};
struct Screen {
    SDL_Window* window = nullptr;
    SDL_Renderer* renderer = nullptr;
    TTF_Font* font = nullptr;
    TTF_Font* title = nullptr;
    SDL_Texture* art = nullptr;
    std::vector<SDL_GameController*> controllers;
    ~Screen() {
        for (auto* controller : controllers) SDL_GameControllerClose(controller);
        if (art) SDL_DestroyTexture(art);
        if (title) TTF_CloseFont(title);
        if (font) TTF_CloseFont(font);
        if (renderer) SDL_DestroyRenderer(renderer);
        if (window) SDL_DestroyWindow(window);
        TTF_Quit(); IMG_Quit(); SDL_Quit();
    }
    void controller(int index) {
        if (SDL_IsGameController(index)) {
            for (auto* pad : controllers)
                if (SDL_JoystickInstanceID(SDL_GameControllerGetJoystick(pad)) == SDL_JoystickGetDeviceInstanceID(index)) return;
            if (auto* pad = SDL_GameControllerOpen(index)) {
                controllers.push_back(pad);
                std::cout << "Controller: " << SDL_GameControllerName(pad) << '\n';
            }
        }
    }
    void text(const std::string& message, int x, int y, int width, SDL_Color color = white, bool heading = false, int max_height = 60) {
        if (message.empty()) return;
        auto* surface = TTF_RenderUTF8_Blended_Wrapped(heading ? title : font, message.c_str(), color, width);
        if (!surface) return;
        auto* texture = SDL_CreateTextureFromSurface(renderer, surface);
        SDL_Rect source{0, 0, surface->w, std::min(surface->h, max_height)};
        SDL_Rect box{x, y, source.w, source.h};
        if (texture) { SDL_RenderCopy(renderer, texture, &source, &box); SDL_DestroyTexture(texture); }
        SDL_FreeSurface(surface);
    }
    void box(SDL_Rect rect, SDL_Color color) {
        SDL_SetRenderDrawColor(renderer, color.r, color.g, color.b, color.a);
        SDL_RenderFillRect(renderer, &rect);
    }
};

Library demo_library() {
    Library result;
    result.games = {{"snes", "Super Nintendo", "demo:snes", "A new adventure awaits", "", true},
                    {"megadrive", "Mega Drive", "demo:md", "One more level", "", false},
                    {"gba", "Game Boy Advance", "demo:gba", "The pocket-sized quest", "", true}};
    return result;
}

// Read image dimensions before asking SDL_image to allocate decoded pixels.
// The first test build accepts PNG/JPEG only, capped at 2048x2048 and 8 MiB.
bool safe_art(const fs::path& path) {
    std::error_code error;
    if (fs::file_size(path, error) > 8 * 1024 * 1024 || error) return false;
    std::ifstream file(path, std::ios::binary);
    unsigned char header[24]{};
    if (!file.read(reinterpret_cast<char*>(header), sizeof(header))) return false;
    const auto bounded = [](unsigned long w, unsigned long h) { return w && h && w <= 2048 && h <= 2048; };
    const unsigned char png[]{137, 80, 78, 71, 13, 10, 26, 10};
    if (std::equal(std::begin(png), std::end(png), header)) {
        auto integer = [&](int offset) {
            return (static_cast<unsigned long>(header[offset]) << 24) | (static_cast<unsigned long>(header[offset + 1]) << 16) |
                   (static_cast<unsigned long>(header[offset + 2]) << 8) | header[offset + 3];
        };
        return bounded(integer(16), integer(20));
    }
    if (header[0] != 255 || header[1] != 216) return false;
    file.clear(); file.seekg(2);
    while (file) {
        if (file.get() != 255) return false;
        int marker;
        do { marker = file.get(); } while (marker == 255);
        if (marker < 0 || marker == 218 || marker == 217) return false;
        const int high = file.get(), low = file.get();
        if (high < 0 || low < 0) return false;
        const int length = (high << 8) | low;
        if (length < 2) return false;
        if (marker >= 192 && marker <= 207 && marker != 196 && marker != 200 && marker != 204) {
            if (length < 8 || file.get() < 0) return false;
            const int hh = file.get(), hl = file.get(), wh = file.get(), wl = file.get();
            return hh >= 0 && hl >= 0 && wh >= 0 && wl >= 0 && bounded((wh << 8) | wl, (hh << 8) | hl);
        }
        file.seekg(length - 2, std::ios::cur);
    }
    return false;
}
}

int main(int argc, char** argv) {
    try {
        bool demo = false, windowed = false;
        int width = 1280, height = 720;
        fs::path screenshot;
        Locations locations;
        fs::path assets = fs::canonical("/proc/self/exe").parent_path() / "assets";
        const char* state_env = std::getenv("PICKER_STATE_DIR");
        fs::path state = state_env ? state_env : "/userdata/system/configs/random-game-picker";
        const char* controller_layout = std::getenv("PICKER_CONTROLLER_LAYOUT");
        const bool knulli_buttons = controller_layout && std::string(controller_layout) == "knulli";
        for (int i = 1; i < argc; ++i) {
            const std::string argument = argv[i];
            if (argument == "--version") { std::cout << "0.1.1\n"; return 0; }
            else if (argument == "--demo") demo = true;
            else if (argument == "--windowed") windowed = true;
            else if (argument == "--assets" && i + 1 < argc) assets = argv[++i];
            else if (argument == "--es-config" && i + 1 < argc) locations.user = argv[++i];
            else if (argument == "--size" && i + 1 < argc) {
                const std::string size = argv[++i];
                const auto split = size.find('x');
                if (split == std::string::npos) throw std::runtime_error("Expected WIDTHxHEIGHT.");
                width = std::stoi(size.substr(0, split)); height = std::stoi(size.substr(split + 1));
                if (width < 640 || height < 480 || width > 3840 || height > 2160) throw std::runtime_error("Unsupported size.");
            } else if (argument == "--screenshot" && i + 1 < argc) screenshot = argv[++i];
            else throw std::runtime_error("Options: --version --demo --windowed --size WIDTHxHEIGHT --assets DIR --es-config DIR --screenshot PNG");
        }
        std::cout << "Controller layout: " << (knulli_buttons ? "KNULLI default ES labels" : "SDL / KNULLI Xbox layout") << '\n';
        locations.home = std::getenv("HOME") ? std::getenv("HOME") : "/userdata/system";
        std::atomic<bool> cancel{false};
        std::future<Library> loading;
        auto scan = [&] {
            return discover(locations, [&](const std::string& route) {
                if (cancel) throw std::runtime_error("Scan cancelled.");
                return es_request("GET", route);
            });
        };
        if (!demo) loading = std::async(std::launch::async, scan);
        Library library = demo ? demo_library() : Library{};
        Json settings = Json::object();
        if (!demo) {
            std::error_code error;
            if (fs::file_size(state / "settings.json", error) <= 32768 && !error) {
                std::ifstream stream(state / "settings.json");
                settings = Json::parse(stream, nullptr, false);
                if (!settings.is_object()) settings = Json::object();
            }
        }
        std::string system = settings.contains("system") && settings["system"].is_string() ? settings["system"].get<std::string>() : "";
        bool favorites = settings.contains("favorites") && settings["favorites"] == true, exclude_recent = false;
        std::vector<std::string> recent;
        std::vector<std::pair<std::string, std::string>> systems{{"", "All Systems"}};
        std::vector<std::size_t> pool;
        std::size_t system_index = 0;
        std::optional<std::size_t> selected;
        bool ready = demo, preview = true, running = true, submitting = false;
        int focus = 0;
        std::string notice = demo ? "Demo library - launching is disabled." : "Reading EmulationStation's library...";
        std::string launch_path;
        std::future<Library> verifying;
        std::mt19937 random(std::random_device{}());
        auto update_pool = [&] { pool = eligible(library, system, favorites, exclude_recent ? recent : std::vector<std::string>{}); };
        auto update_systems = [&] {
            systems.resize(1);
            std::set<std::string> seen;
            for (const auto& game : library.games)
                if (seen.insert(game.system).second) systems.emplace_back(game.system, game.system_label);
            system_index = 0;
            for (std::size_t i = 0; i < systems.size(); ++i) if (systems[i].first == system) system_index = i;
            system = systems[system_index].first;
            update_pool();
        };
        // Keep the saved system until discovery can resolve it against real games.
        if (ready) update_systems();
        {
            Screen screen;
            if (SDL_Init(SDL_INIT_VIDEO | SDL_INIT_GAMECONTROLLER) != 0 || TTF_Init() != 0) throw std::runtime_error(SDL_GetError());
            SDL_GameControllerAddMappingsFromFile("/tmp/gamecontrollerdb.txt");
            IMG_Init(IMG_INIT_PNG | IMG_INIT_JPG);
            screen.window = SDL_CreateWindow("KNULLI Random Game Picker", SDL_WINDOWPOS_CENTERED, SDL_WINDOWPOS_CENTERED,
                                              width, height, windowed ? SDL_WINDOW_SHOWN : SDL_WINDOW_FULLSCREEN_DESKTOP);
            if (!screen.window) throw std::runtime_error(SDL_GetError());
            SDL_GetWindowSize(screen.window, &width, &height);
            screen.renderer = SDL_CreateRenderer(screen.window, -1, SDL_RENDERER_ACCELERATED | SDL_RENDERER_PRESENTVSYNC);
            if (!screen.renderer) screen.renderer = SDL_CreateRenderer(screen.window, -1, SDL_RENDERER_SOFTWARE);
            if (!screen.renderer) throw std::runtime_error(SDL_GetError());
            // A 640x480 logical canvas keeps text readable on the smaller handheld.
            const bool wide = width * 3 >= height * 5;
            const int canvas_width = wide ? 854 : 640;
            SDL_RenderSetLogicalSize(screen.renderer, canvas_width, 480);
            screen.font = TTF_OpenFont((assets / "DejaVuSans.ttf").c_str(), 18);
            screen.title = TTF_OpenFont((assets / "DejaVuSans.ttf").c_str(), 28);
            if (!screen.font || !screen.title) throw std::runtime_error("Bundled font missing. Reinstall the complete archive.");
            for (int i = 0; i < SDL_NumJoysticks(); ++i) screen.controller(i);
            auto roll = [&] {
                update_pool();
                selected = choose(pool, random);
                preview = true;
                notice.clear();
                recent.push_back(library.games[*selected].path);
                if (recent.size() > 10) recent.erase(recent.begin());
                if (screen.art) { SDL_DestroyTexture(screen.art); screen.art = nullptr; }
                // Load only the selected image. Large/corrupt files fall back to text.
                const auto& image = library.games[*selected].image;
                if (!image.empty() && safe_art(image)) {
                    SDL_Surface* surface = IMG_Load(image.c_str());
                    if (surface) {
                        if (surface->w <= 2048 && surface->h <= 2048) screen.art = SDL_CreateTextureFromSurface(screen.renderer, surface);
                        SDL_FreeSurface(surface);
                    }
                }
            };
            if (demo) roll();
            while (running) {
                if (loading.valid() && loading.wait_for(std::chrono::seconds(0)) == std::future_status::ready) {
                    try {
                        library = loading.get(); ready = true; update_systems();
                        if (preview && !pool.empty()) roll();
                        notice = library.games.empty() ? "No eligible games. Check ES library and system configuration." :
                                 (library.warnings.empty() ? "" : library.warnings.front());
                    } catch (const std::exception& error) { notice = error.what(); }
                }
                if (verifying.valid() && verifying.wait_for(std::chrono::seconds(0)) == std::future_status::ready) {
                    try {
                        const auto fresh = verifying.get();
                        const auto& game = library.games[*selected];
                        const auto match = std::find_if(fresh.games.begin(), fresh.games.end(), [&](const Game& item) {
                            return item.path == game.path && item.system == game.system;
                        });
                        if (match == fresh.games.end()) throw std::runtime_error("Selected game is no longer launchable. Rescan from Settings.");
                        launch_path = game.path;
                        running = false;
                    } catch (const std::exception& error) { notice = error.what(); submitting = false; }
                }
                SDL_Event event;
                while (SDL_PollEvent(&event)) {
                    if (event.type == SDL_QUIT) { running = false; cancel = true; launch_path.clear(); }
                    if (event.type == SDL_CONTROLLERDEVICEADDED) screen.controller(event.cdevice.which);
                    if (event.type == SDL_CONTROLLERDEVICEREMOVED) {
                        auto& pads = screen.controllers;
                        pads.erase(std::remove_if(pads.begin(), pads.end(), [](SDL_GameController* pad) {
                            if (SDL_GameControllerGetAttached(pad)) return false;
                            SDL_GameControllerClose(pad); return true;
                        }), pads.end());
                    }
                    int action = 0;
                    if (event.type == SDL_KEYUP) {
                        switch (event.key.keysym.sym) {
                            case SDLK_RETURN: action = 'a'; break;
                            case SDLK_ESCAPE: action = 'b'; break;
                            case SDLK_x: action = 'x'; break;
                            case SDLK_y: action = 'y'; break;
                            case SDLK_UP: action = 'u'; break;
                            case SDLK_DOWN: action = 'd'; break;
                            case SDLK_LEFT: action = 'l'; break;
                            case SDLK_RIGHT: action = 'r'; break;
                        }
                    }
                    if (event.type == SDL_CONTROLLERBUTTONUP) {
                        auto button = event.cbutton.button;
                        // Configgen maps ES a/b/x/y to SDL b/a/y/x, except in Xbox-layout mode.
                        // Translate back to the ES labels printed by this UI, not raw device IDs.
                        if (knulli_buttons) {
                            switch (button) {
                                case SDL_CONTROLLER_BUTTON_A: button = SDL_CONTROLLER_BUTTON_B; break;
                                case SDL_CONTROLLER_BUTTON_B: button = SDL_CONTROLLER_BUTTON_A; break;
                                case SDL_CONTROLLER_BUTTON_X: button = SDL_CONTROLLER_BUTTON_Y; break;
                                case SDL_CONTROLLER_BUTTON_Y: button = SDL_CONTROLLER_BUTTON_X; break;
                            }
                        }
                        switch (button) {
                            case SDL_CONTROLLER_BUTTON_A: action = 'a'; break;
                            case SDL_CONTROLLER_BUTTON_B: action = 'b'; break;
                            case SDL_CONTROLLER_BUTTON_X: action = 'x'; break;
                            case SDL_CONTROLLER_BUTTON_Y: action = 'y'; break;
                            case SDL_CONTROLLER_BUTTON_DPAD_UP: action = 'u'; break;
                            case SDL_CONTROLLER_BUTTON_DPAD_DOWN: action = 'd'; break;
                            case SDL_CONTROLLER_BUTTON_DPAD_LEFT: action = 'l'; break;
                            case SDL_CONTROLLER_BUTTON_DPAD_RIGHT: action = 'r'; break;
                            case SDL_CONTROLLER_BUTTON_START: action = 'q'; break;
                        }
                    }
                    if (!action || submitting) continue;
                    try {
                        if (action == 'q') { running = false; cancel = true; }
                        else if (action == 'b') { if (preview) preview = false; else { running = false; cancel = true; } }
                        else if (action == 'y' && preview) { preview = false; notice.clear(); }
                        else if (action == 'y' && !loading.valid()) {
                            if (demo) { notice = "Demo mode: no device scan or launch."; continue; }
                            selected.reset(); ready = false; notice = "Rescanning EmulationStation...";
                            loading = std::async(std::launch::async, scan);
                        } else if (ready && preview) {
                            if (action == 'x') roll();
                            if (action == 'a' && selected) {
                                if (demo) { notice = "Demo only - no launch request was sent."; continue; }
                                notice = "Checking launch eligibility..."; submitting = true;
                                verifying = std::async(std::launch::async, scan);
                            }
                        } else if (ready) {
                            if (action == 'u') focus = (focus + 3) % 4;
                            if (action == 'd') focus = (focus + 1) % 4;
                            if (action == 'l' || action == 'r' || action == 'a') {
                                if (focus == 0) {
                                    if (action == 'a') roll();
                                    else {
                                        system_index = action == 'r' ? (system_index + 1) % systems.size() :
                                                                 (system_index + systems.size() - 1) % systems.size();
                                        system = systems[system_index].first; update_pool();
                                    }
                                } else if (focus == 1) { favorites = !favorites; update_pool(); }
                                else if (focus == 2) { exclude_recent = !exclude_recent; update_pool(); }
                                else { recent.clear(); update_pool(); notice = "Recent selections cleared."; }
                            }
                        }
                    } catch (const std::exception& error) { notice = error.what(); }
                }
                screen.box({0, 0, canvas_width, 480}, {12, 22, 30, 255});
                screen.text("RANDOM GAME PICKER", 28, 20, canvas_width - 56, white, true);
                screen.text(demo ? "DEMO / NO LAUNCH" : "v0.1.1", 28, 61, canvas_width - 56, accent);
                if (!preview) {
                    const std::vector<std::string> rows{
                        "System:  " + systems[system_index].second,
                        std::string("Favorites only:  ") + (favorites ? "ON" : "OFF"),
                        std::string("Exclude last 10 picks:  ") + (exclude_recent ? "ON" : "OFF"),
                        "Reset recent selections"};
                    for (int i = 0; i < 4; ++i) {
                        const int y = 107 + i * 47;
                        screen.box({24, y, canvas_width - 48, 40}, i == focus ? SDL_Color{29, 66, 65, 255} : SDL_Color{22, 36, 47, 255});
                        screen.text(rows[i], 38, y + 8, canvas_width - 76, i == focus ? accent : white);
                    }
                    screen.text(std::to_string(pool.size()) + " eligible games", 28, 308, canvas_width - 56, muted);
                    screen.text("A: Pick / toggle    Left/Right: system", 28, 397, canvas_width - 56);
                } else if (selected) {
                    const auto& game = library.games[*selected];
                    SDL_Rect image{28, 107, wide ? 290 : 185, 225};
                    screen.box(image, {22, 36, 47, 255});
                    if (screen.art) {
                        int w, h; SDL_QueryTexture(screen.art, nullptr, nullptr, &w, &h);
                        const double scale = std::min(double(image.w) / w, double(image.h) / h);
                        SDL_Rect fit{image.x + (image.w - int(w * scale)) / 2, image.y + (image.h - int(h * scale)) / 2, int(w * scale), int(h * scale)};
                        SDL_RenderCopy(screen.renderer, screen.art, nullptr, &fit);
                    } else screen.text("NO ARTWORK\n\nYour scraped art\nwill appear here", image.x + 12, image.y + 48, image.w - 24, muted, false, 160);
                    const int x = image.x + image.w + 24;
                    screen.text(game.title, x, 110, canvas_width - x - 28, white, true, 150);
                    screen.text(game.system_label + (game.favorite ? "\nFavorite" : ""), x, 275, canvas_width - x - 28, accent);
                    screen.text("A: Play    X: Reroll    B: Filters", 28, 397, canvas_width - 56);
                } else {
                    screen.text(ready ? "NO ELIGIBLE GAMES" : (loading.valid() ? "READING LIBRARY" : "LIBRARY UNAVAILABLE"),
                                28, 130, canvas_width - 56, white, true);
                    screen.text("Change filters or rescan from Settings.", 28, 190, canvas_width - 56, muted);
                    screen.text("B: Filters", 28, 397, canvas_width - 56);
                }
                screen.text(notice, 28, 344, canvas_width - 56, accent, false, 44);
                screen.text(preview ? "Y: Settings    Start: Quit" : "Y: Rescan    B: Exit    Start: Quit", 28, 438, canvas_width - 56, muted);
                if (!screenshot.empty() && ready) {
                    SDL_Surface* frame = SDL_CreateRGBSurfaceWithFormat(0, width, height, 32, SDL_PIXELFORMAT_RGBA32);
                    if (frame) {
                        SDL_RenderReadPixels(screen.renderer, nullptr, frame->format->format, frame->pixels, frame->pitch);
                        IMG_SavePNG(frame, screenshot.c_str()); SDL_FreeSurface(frame);
                    }
                    screenshot.clear();
                }
                SDL_RenderPresent(screen.renderer);
                SDL_Delay(16);
            }
        }
        cancel = true;
        if (!demo) {
            fs::create_directories(state);
            const auto temporary = state / "settings.json.tmp";
            { std::ofstream file(temporary); file << Json{{"system", system}, {"favorites", favorites}}.dump(2); }
            fs::rename(temporary, state / "settings.json");
        }
        // All SDL/input resources and the UI loop are closed before handoff.
        if (!launch_path.empty()) {
            std::cout << "Submitting one ES launch request; HTTP acceptance is not confirmation of play.\n";
            es_request("POST", "/launch", launch_path);
        }
        return 0;
    } catch (const std::exception& error) {
        std::cerr << "Random Game Picker: " << error.what() << '\n';
        return 1;
    }
}
