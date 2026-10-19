#pragma once

#include <filesystem>
#include <functional>
#include <nlohmann/json.hpp>
#include <random>
#include <string>
#include <vector>

namespace picker {
namespace fs = std::filesystem;
using Json = nlohmann::json;
using Fetch = std::function<std::string(const std::string&)>;

struct Game {
    std::string system, system_label, path, title, image;
    bool favorite = false;
};
struct Library {
    std::vector<Game> games;
    std::vector<std::string> warnings;
};
struct Locations {
    fs::path user = "/userdata/system/configs/emulationstation";
    fs::path shared = "/usr/share/emulationstation";
    fs::path legacy = "/etc/emulationstation";
    fs::path home;
};

Library discover(const Locations& locations, const Fetch& fetch);
std::vector<std::size_t> eligible(const Library& library, const std::string& system,
                                bool favorites, const std::vector<std::string>& recent);
std::size_t choose(const std::vector<std::size_t>& pool, std::mt19937& engine);
std::string encode_segment(const std::string& text);
}
