// Test-only input source. Feed logical SDL button presses through the real event queue.
#include <SDL.h>
#include <cstdlib>
#include <dlfcn.h>
#include <fcntl.h>
#include <unistd.h>

static void inject() {
    static const int input = open(std::getenv("PICKER_TEST_INPUT_FIFO"), O_RDONLY | O_NONBLOCK);
    unsigned char button;
    if (read(input, &button, 1) == 1) {
        SDL_Event release{};
        release.type = SDL_CONTROLLERBUTTONDOWN;
        release.cbutton.state = SDL_PRESSED;
        release.cbutton.button = button;
        SDL_PushEvent(&release);
        release.type = SDL_CONTROLLERBUTTONUP;
        release.cbutton.state = SDL_RELEASED;
        release.cbutton.button = button;
        SDL_PushEvent(&release);
    }
}

extern "C" int SDL_PollEvent(SDL_Event* event) {
    static auto poll = reinterpret_cast<int (*)(SDL_Event*)>(dlsym(RTLD_NEXT, "SDL_PollEvent"));
    inject();
    return poll(event);
}

extern "C" int SDL_WaitEventTimeout(SDL_Event* event, int timeout) {
    static auto wait = reinterpret_cast<int (*)(SDL_Event*, int)>(dlsym(RTLD_NEXT, "SDL_WaitEventTimeout"));
    inject();
    return wait(event, timeout);
}
