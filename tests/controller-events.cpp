// Test-only input source. Feed logical SDL button releases through the real event queue.
#include <SDL.h>
#include <cstdlib>
#include <dlfcn.h>
#include <fcntl.h>
#include <unistd.h>

extern "C" int SDL_PollEvent(SDL_Event* event) {
    static auto poll = reinterpret_cast<int (*)(SDL_Event*)>(dlsym(RTLD_NEXT, "SDL_PollEvent"));
    static const int input = open(std::getenv("PICKER_TEST_INPUT_FIFO"), O_RDONLY | O_NONBLOCK);
    unsigned char button;
    if (read(input, &button, 1) == 1) {
        SDL_Event release{};
        release.type = SDL_CONTROLLERBUTTONUP;
        release.cbutton.button = button;
        SDL_PushEvent(&release);
    }
    return poll(event);
}
