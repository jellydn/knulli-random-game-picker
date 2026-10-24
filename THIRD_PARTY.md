# Third-party notices

The application is MIT licensed. The unreleased Go rewrite uses locked modules
listed in `go.mod`/`go.sum`, including gabagool/v2 and go-sdl2 (MIT), Go libraries
(BSD-style), and their transitive dependencies. `scripts/package-device` includes
each downloaded module's LICENSE/COPYING/NOTICE files and the DejaVu font notice.
Build tags omit gabagool's embedded fonts; the UI uses the packaged DejaVu Sans.
The Go toolchain license is also included (https://go.dev/LICENSE).

The new runtime dynamically requires SDL2_gfx as well as SDL2/image/ttf and glibc.
These libraries are supplied by firmware, not bundled. Inspect `abi.txt` before use.
No Grout source, ROM, commercial artwork or firmware image is bundled.

## Historical C++ release 0.1.1

Published 0.1.1 packages include original distribution copyright/license notices for:

- pugixml (MIT), statically linked: https://pugixml.org/
- nlohmann/json (MIT), compiled headers: https://github.com/nlohmann/json
- GCC libstdc++/libgcc (GPL with GCC Runtime Library Exception), statically linked: https://gcc.gnu.org/onlinedocs/libstdc++/manual/license.html
- DejaVu Sans (Bitstream Vera-derived font license), bundled unmodified: https://dejavu-fonts.github.io/

The executable dynamically uses firmware-provided SDL2, SDL2_image, SDL2_ttf (zlib licenses), and glibc including libm and the ELF loader (LGPL). These system libraries are not distributed in the ZIP or modified by the application. Their transitive firmware dependencies remain supplied by KNULLI. See the published `abi.txt` for the executable's exact interpreter and DT_NEEDED entries.

Build dependencies come from Debian Bookworm; exact GPL library source and the Runtime Library Exception are available through Debian's GCC source package and https://www.gnu.org/licenses/gcc-exception-3.1.html. The runtime exception permits linking this application under its own license. No commercial ROM, scraped artwork, or firmware image is bundled.
