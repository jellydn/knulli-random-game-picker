#include "http.hpp"
#include <iostream>
#include <stdexcept>
int main(int argc, char** argv) {
    if (argc != 4) return 2;
    try {
        std::cout << picker::es_request(argv[1], argv[2], argv[3]);
        return 0;
    } catch (const std::exception& error) {
        std::cerr << error.what();
        return 1;
    }
}
