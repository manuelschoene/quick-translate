# Changelog

## 0.1.0-alpha (2026-09-11)


### Features

* Added CI, issue templates and docs ([c53ca76](https://github.com/manuelschoene/quick-translate/commit/c53ca7632bfd8423eca9cdeca420b053f54b0124))
* Added history package for keeping track of past translations ([0eb9821](https://github.com/manuelschoene/quick-translate/commit/0eb98214e9f84f1f77f6a88fb14ff1ca1f175472))
* Added keyboard navigation in language selection and failure animation to buttons ([0a57ded](https://github.com/manuelschoene/quick-translate/commit/0a57dede7866b152af04c6b7ec54de3ca48d092a))
* Added Linux and KDE integration ([8db8ffe](https://github.com/manuelschoene/quick-translate/commit/8db8ffe7b7d8163dd4ccaf53fb058f892b86e2ed))
* Added text translation via DeepL ([5934ae4](https://github.com/manuelschoene/quick-translate/commit/5934ae4ef9ca85161517fcb59a1f46056b62d3f6))
* Added X11 clipboard support ([a1647ac](https://github.com/manuelschoene/quick-translate/commit/a1647ac96bb016ab324624275f974af5b32a2ab6))
* Allow configuration of provider via config file ([e0cd5bd](https://github.com/manuelschoene/quick-translate/commit/e0cd5bdbf6cc1b9f4909b0f06c7b834ef78ce80d))
* Enabled language retrieval ([dbd3f49](https://github.com/manuelschoene/quick-translate/commit/dbd3f49b129c5c4921c68625bca89bff70e670ae))
* Implemented core package ([60f2ca3](https://github.com/manuelschoene/quick-translate/commit/60f2ca33e5a059d9f69a80c5b272d34900f67aa6))
* Implemented data abstraction layer in frontend ([5602242](https://github.com/manuelschoene/quick-translate/commit/5602242896f2400e77ffe02c84f28fc1bcda5b70))
* Implemented error view ([5a96cdc](https://github.com/manuelschoene/quick-translate/commit/5a96cdcc7d8b8d91abc24ecaaa5ed71fcbb989b5))
* Implemented IPC ([f2064fe](https://github.com/manuelschoene/quick-translate/commit/f2064fe508868890339a2f1c35916bed65d0de98))
* Implemented language package ([48ed625](https://github.com/manuelschoene/quick-translate/commit/48ed6252d652dc7e8595adbcc552270e6349fa6b))
* Implemented language selection views ([501e514](https://github.com/manuelschoene/quick-translate/commit/501e5148dd0a05562983562d9696c6bcde9c126c))
* Implemented translation view ([cc91908](https://github.com/manuelschoene/quick-translate/commit/cc91908898b9f6a04436c73be20c1a0b719650d6))
* Implemented transport layer ([562d082](https://github.com/manuelschoene/quick-translate/commit/562d08256fa676e1b67493749be4d39986509385))
* Improved Linux support and installation pipeline ([51791cb](https://github.com/manuelschoene/quick-translate/commit/51791cb3e62798334983e8bb5bd35e33c50bb7dc))
* Integrated clipboard reading and writing ([87fc97f](https://github.com/manuelschoene/quick-translate/commit/87fc97fea82ebca52052d711c28979a672212fba))
* Reworked clipboard package for better OS support. ([e4a1659](https://github.com/manuelschoene/quick-translate/commit/e4a165919ba33bc340a31771b6244f1a856c5f86))
* Sending notification with translation ([2abd377](https://github.com/manuelschoene/quick-translate/commit/2abd3777e972f59280334630616733d4e50f2306))


### Bug Fixes

* Added missing builder setter for detected source language. ([a19ab58](https://github.com/manuelschoene/quick-translate/commit/a19ab58f04139c9965cd24e514d4ae1c03e31c62))
* Fixed autofocus on input in SearchBox component ([1e21049](https://github.com/manuelschoene/quick-translate/commit/1e2104955e379e949790eeef75abcd4a6f79d8dd))
* Fixed bugs in language package ([9be6ade](https://github.com/manuelschoene/quick-translate/commit/9be6adee0dc930208e799da8ae20a58b30e46533))
* Fixed CI pipeline by adding missing Wails bindings ([dfef5f0](https://github.com/manuelschoene/quick-translate/commit/dfef5f0cdc43d64b9ab16d8a1b6d526ad8ce4de0))
* Fixed crash-loop, missing timeout and file permissions ([05e685f](https://github.com/manuelschoene/quick-translate/commit/05e685f2b856d9d5b2be6ec74f852aaa613110b9))
* Fixed inconsistencies with ESLint ([851c302](https://github.com/manuelschoene/quick-translate/commit/851c3028589fa3bf9716140775c8970a7b7b729c))
* Fixed invalid path in systemd service file ([e3db082](https://github.com/manuelschoene/quick-translate/commit/e3db082fb6abbcfcf06d96b0e3f6d9210553bc64))
* Fixed systemd service registration failing in sandboxes ([ac13fd2](https://github.com/manuelschoene/quick-translate/commit/ac13fd2d304ff554e7478ff2f076702f781f7a11))
* Fixed vulnerabilities by bumping Go version ([39abe41](https://github.com/manuelschoene/quick-translate/commit/39abe4114844a07db5d23cae3a81df07d64234b9))
