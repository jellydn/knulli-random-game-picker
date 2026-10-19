# Third-party notices

The application is MIT licensed. Release packages include the original distribution copyright/license notices in `licenses/` for:

- pugixml (MIT), statically linked: https://pugixml.org/
- nlohmann/json (MIT), compiled headers: https://github.com/nlohmann/json
- GCC libstdc++/libgcc (GPL with GCC Runtime Library Exception), statically linked: https://gcc.gnu.org/onlinedocs/libstdc++/manual/license.html
- DejaVu Sans (Bitstream Vera-derived font license), bundled unmodified: https://dejavu-fonts.github.io/

The executable dynamically uses firmware-provided SDL2, SDL2_image, SDL2_ttf (zlib licenses), and glibc including libm and the ELF loader (LGPL). These system libraries are not distributed in the ZIP or modified by the application. Their transitive firmware dependencies remain supplied by KNULLI. See the published `abi.txt` for the executable's exact interpreter and DT_NEEDED entries.

Build dependencies come from Debian Bookworm; exact GPL library source and the Runtime Library Exception are available through Debian's GCC source package and https://www.gnu.org/licenses/gcc-exception-3.1.html. The runtime exception permits linking this application under its own license. No commercial ROM, scraped artwork, or firmware image is bundled.
