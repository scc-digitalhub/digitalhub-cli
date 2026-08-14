## [0.15.2](https://github.com/scc-digitalhub/digitalhub-cli/compare/0.15.1...0.15.2) (2026-08-14)

### Bug Fixes

* configuration refresh + writing to ini ordering issues ([fde5160](https://github.com/scc-digitalhub/digitalhub-cli/commit/fde516026906a6f1bbbafc7b714269c1ec6bf576))
* drop custom logic for picking containers for logs ([7f39135](https://github.com/scc-digitalhub/digitalhub-cli/commit/7f391354c15741afe268912a3ae941411d512dc7))
* move keys to package to separate and cleanup cross imports ([45cb27e](https://github.com/scc-digitalhub/digitalhub-cli/commit/45cb27ed6721e3aef278a259b09949e68bb12ac4))
* register always sets the new one as default ([09a23df](https://github.com/scc-digitalhub/digitalhub-cli/commit/09a23dfc1a626b17b65e890170fddb2176cefd59))

### Features

* add provider filter to config and credentials export, with custom exporters for well-known ([e6477c8](https://github.com/scc-digitalhub/digitalhub-cli/commit/e6477c86cc8235fffa3438c3e2f2d88c9957736f))
* auto-refresh stale credentials when possible ([e3a1fb0](https://github.com/scc-digitalhub/digitalhub-cli/commit/e3a1fb0439668d32bc627eb5b1eb9891a955ca30))
* cli metrics ([cd964fd](https://github.com/scc-digitalhub/digitalhub-cli/commit/cd964fd5097940f7d90dcdc67823473c2adc2747))
* listenv shows current env ([b5150e7](https://github.com/scc-digitalhub/digitalhub-cli/commit/b5150e7a193c9580da58eb1bce19c5e9a3efa036))
* remove hardcoded list of config/cred keys and properly handle source/dest ([60cf142](https://github.com/scc-digitalhub/digitalhub-cli/commit/60cf14218dee7148085974a8ce56031f4834f4dc))
* support and explicit use oauth2 token expiration ([25a26fe](https://github.com/scc-digitalhub/digitalhub-cli/commit/25a26fe81805505360e4da82a4b26fc47efe9072))
* support events streaming from core ([ef3db97](https://github.com/scc-digitalhub/digitalhub-cli/commit/ef3db97e3389d196a1233be6c3b23471ca56ccf1))
* support local ini file before falling back to homedir ([c74a679](https://github.com/scc-digitalhub/digitalhub-cli/commit/c74a679e75aabf6772ef596729245814e38e5fd8))

## [0.15.1](https://github.com/scc-digitalhub/digitalhub-cli/compare/0.15.0...0.15.1) (2026-07-27)

### Bug Fixes

* update env refresh logic to handle partial configs via api_level and read-only config files ([b1492ed](https://github.com/scc-digitalhub/digitalhub-cli/commit/b1492edc9bf273da824819b1019ee74c25862bf3))

### Features

* non-interactive login via PAT ([a9e1bb8](https://github.com/scc-digitalhub/digitalhub-cli/commit/a9e1bb8acf379b29af385e27f07c43d25c323006))
* venv init handle errors + supports uv ([57a8113](https://github.com/scc-digitalhub/digitalhub-cli/commit/57a81135e385df68c524ad331da4e9e244fcb241))

# [0.15.0](https://github.com/scc-digitalhub/digitalhub-cli/compare/0.15.0-beta3...0.15.0) (2026-07-17)

### Bug Fixes

* update sdk to fix upload issues ([e9c35e7](https://github.com/scc-digitalhub/digitalhub-cli/commit/e9c35e71f73e33f5b2babd32cc00e61bcc979046))

### Features

* add resolvers for run id from function name and run name for proxy and pf ([92018ae](https://github.com/scc-digitalhub/digitalhub-cli/commit/92018ae8762472c671fec260332006533bbd6f55))
* proxy for browser access ([a579acb](https://github.com/scc-digitalhub/digitalhub-cli/commit/a579acbdfabfa2290b9069c7d695acc1beb9b61a))

# [0.15.0-beta3](https://github.com/scc-digitalhub/digitalhub-cli/compare/0.15.0-beta2...0.15.0-beta3) (2026-05-04)

# [0.15.0-beta2](https://github.com/scc-digitalhub/digitalhub-cli/compare/0.15.0-beta1...0.15.0-beta2) (2026-05-04)

### Features

* proxy uses X-proxy to enable TLS transport ([ac13987](https://github.com/scc-digitalhub/digitalhub-cli/commit/ac139871b045e4043f924787e2a2ae67c4f845c8))

# [0.15.0-beta1](https://github.com/scc-digitalhub/digitalhub-cli/compare/0.14.4...0.15.0-beta1) (2026-04-17)

### Bug Fixes

* proxy env var for PROJECT_NAME ([0a0c4d9](https://github.com/scc-digitalhub/digitalhub-cli/commit/0a0c4d9159a803f1ab1ea14a804945eb3fb8bf94))

### Features

* remove deprecated resume command ([10c78ec](https://github.com/scc-digitalhub/digitalhub-cli/commit/10c78ec48534b811fe3efa581f8b91462ec44e9e))
* transparent http proxy support for run ([1d47126](https://github.com/scc-digitalhub/digitalhub-cli/commit/1d4712674e9b1968dcd74d9eca12f4117b6e7cfc))

## [0.14.4](https://github.com/scc-digitalhub/digitalhub-cli/compare/0.14.3...0.14.4) (2026-04-16)

### Bug Fixes

* log/debug msg to stderr ([c470b41](https://github.com/scc-digitalhub/digitalhub-cli/commit/c470b419a20fedaf5c7a4eec488905b42443e16f))
* login and refresh update all credentials back to ini ([e6dd147](https://github.com/scc-digitalhub/digitalhub-cli/commit/e6dd147ce68ccba2be687ba1a6c52d41c2789ab4))
* new login server logic to fix issues with port and browser/os ([e997c57](https://github.com/scc-digitalhub/digitalhub-cli/commit/e997c5754ca2dc33e480aeb694b2fc6477a2213f))
* STOP is meaningful only for runs ([7601844](https://github.com/scc-digitalhub/digitalhub-cli/commit/7601844f99186c87b90e907d394c9c3fa8169775))

### Features

* add support for PROJECT_NAME env for selecting project ([f57e3c3](https://github.com/scc-digitalhub/digitalhub-cli/commit/f57e3c34c01e2f2b5f6cd0186dedd5994d745d11))
* callback.html proper template ([740e646](https://github.com/scc-digitalhub/digitalhub-cli/commit/740e646112e65e97e9e9e15f622d0fa4787c076d))
* check before register for conflicts, expose --force flag ([95f136e](https://github.com/scc-digitalhub/digitalhub-cli/commit/95f136edef18a7525ca2aa67ad1565ec4315a853))
* dynamic port for oauth2 callback ([4f1ba6b](https://github.com/scc-digitalhub/digitalhub-cli/commit/4f1ba6b41b28d1587cf23a8dfb636ad9f5980784))
* init command creates venv ([f4c1f94](https://github.com/scc-digitalhub/digitalhub-cli/commit/f4c1f9462566f8618149377f0a4c07ecb5c95753))
* list uses proper tabwriter instead of fixed length ([a0ebe52](https://github.com/scc-digitalhub/digitalhub-cli/commit/a0ebe52c02b1803f6bdd0ed23e7fd0982d5a81c1))
* log follow without clear screen ([b6c4e28](https://github.com/scc-digitalhub/digitalhub-cli/commit/b6c4e283a9846293b4b3d0fc9969066c21d9ba97))
* logger ([0483e25](https://github.com/scc-digitalhub/digitalhub-cli/commit/0483e25cb8e63cd6e2bfcabb5852c80a49ba32c6))
* new sdk version + fix golang to 1.24 ([803ab66](https://github.com/scc-digitalhub/digitalhub-cli/commit/803ab669a143e6d69ed1736b9b375c4599e1cc87))
* services command, defaults to RUNNING ([efad823](https://github.com/scc-digitalhub/digitalhub-cli/commit/efad823fe2ac1876abfef7e6029d87116a140738))
* verbose and debug modes with full http trace ([d208f7a](https://github.com/scc-digitalhub/digitalhub-cli/commit/d208f7a3e46d8ce59395eb6359314ce2cd60937a))

## [0.14.3](https://github.com/scc-digitalhub/digitalhub-cli/compare/0.14.2...0.14.3) (2026-03-31)

### Bug Fixes

* fix version and goreleaser file, now when new release is done, it should create also a file in Formula folder for brew package installation ([4cc53c1](https://github.com/scc-digitalhub/digitalhub-cli/commit/4cc53c17e7a1b77a87f1bbb561d6747cec2ff00c))

### Features

* add credentials and config command ([6037e9c](https://github.com/scc-digitalhub/digitalhub-cli/commit/6037e9c9d35bfb0fd1308c2f3ab87da1033aa30d))

## [0.14.2](https://github.com/scc-digitalhub/digitalhub-cli/compare/0.14.1...0.14.2) (2025-12-17)

### Bug Fixes

* Add missed file globalprogress ([6e666c8](https://github.com/scc-digitalhub/digitalhub-cli/commit/6e666c8d8f8957316d0b8c5c68d7d2798f53fe61))
* Delete old service directory ([8a37f75](https://github.com/scc-digitalhub/digitalhub-cli/commit/8a37f7599a2eede4476b46797ce9e18c692e8759))
* **sdk:** add some missed part to create, get, list and run adapters ([b841d65](https://github.com/scc-digitalhub/digitalhub-cli/commit/b841d65494ce95b4e112421ace72ce741db2ef91))
* **sdk:** restore login to be not part of sdk ([935ed8f](https://github.com/scc-digitalhub/digitalhub-cli/commit/935ed8f7c42588f90248a8edc1c25af9909ad578))
* **viper:** add new field ini_source. This prevent to update environment when ini file is created from envs ([d49b993](https://github.com/scc-digitalhub/digitalhub-cli/commit/d49b993223509b0c7ae25104222e9f1ae99a03f9))
* **viper:** when ini file not present read directly from env without taking into account  .wellknown configuration ([b9519a2](https://github.com/scc-digitalhub/digitalhub-cli/commit/b9519a2da7a729cb8c57a1d1a3e2ba89a239321a))

### Features

* add download and upload info with percentage when no --verbose flag is passed to the command ([5dde5d3](https://github.com/scc-digitalhub/digitalhub-cli/commit/5dde5d3ee2ab459bc2366f8e2ff302019c156135))
* extract upload service as SDK ([1d88e1d](https://github.com/scc-digitalhub/digitalhub-cli/commit/1d88e1d9e25aa5670929f1b4c57975cf36fc09b4))
* **login:** move login functionality to sdk ([5a9fc0f](https://github.com/scc-digitalhub/digitalhub-cli/commit/5a9fc0fc358c1c6f3cb6af9636b97a53894c4785))
* **sdk:** add delete/create command to sdk ([ec0c321](https://github.com/scc-digitalhub/digitalhub-cli/commit/ec0c32143267337d1764027fc1d15428f60ceb6a))
* **sdk:** add get command to sdk ([13bc651](https://github.com/scc-digitalhub/digitalhub-cli/commit/13bc65108af31ff14350cb054d94d255c528d325))
* **sdk:** add list and get tests. This show how to use the sdk without pass throught the cli ([9766ba9](https://github.com/scc-digitalhub/digitalhub-cli/commit/9766ba92befd5e3290b9136f3fddd3c7c699faa7))
* **sdk:** add list command to sdk ([3f04104](https://github.com/scc-digitalhub/digitalhub-cli/commit/3f04104cd1c3f17642d7499f413dd27692bbba81))
* **sdk:** add log and metrics to sdk ([22b87de](https://github.com/scc-digitalhub/digitalhub-cli/commit/22b87de4e4bbab9b5fda3fcefbe78e2f9fda7290))
* **sdk:** add run command to sdk ([4b8a595](https://github.com/scc-digitalhub/digitalhub-cli/commit/4b8a595badf01908d858676dc05bab8869f0244f))
* **sdk:** add stop and resume commands to sdk ([c32f0be](https://github.com/scc-digitalhub/digitalhub-cli/commit/c32f0bee3812405ace8412477791e8be3090939b))
* **sdk:** add update command to sdk ([4432d45](https://github.com/scc-digitalhub/digitalhub-cli/commit/4432d45f98b9d501119604bed70e2b3f09bbd9f5))
* **sdk:** align code between main and sdk ([c7fc348](https://github.com/scc-digitalhub/digitalhub-cli/commit/c7fc348d1734ded844f7903a240c3db8cdd90363))
* **sdk:** align sdk version to work as v14 version, with new env variables ([9eb315c](https://github.com/scc-digitalhub/digitalhub-cli/commit/9eb315c4392de813880a09195d7e08e2e293fea1))
* **sdk:** merge services into categories, crud, run, transfer. Main services is now Facade with internal adapter ([8cf087a](https://github.com/scc-digitalhub/digitalhub-cli/commit/8cf087ad5649b73844c8cd1000542e98e1eb63a3))
* **viper:** give precedence to env over ini file ([cc61202](https://github.com/scc-digitalhub/digitalhub-cli/commit/cc612029f5f651c25675aa69553a64179eead352))
* working on list command ([e97916b](https://github.com/scc-digitalhub/digitalhub-cli/commit/e97916b101b5bcdf9c1d047081cd89ec5291573d))

# [0.14.0](https://github.com/scc-digitalhub/digitalhub-cli/compare/0.14.0-beta.6...0.14.0) (2025-10-28)

### Features

* **upload:** add lineage when upload artifact ([d3ad9dd](https://github.com/scc-digitalhub/digitalhub-cli/commit/d3ad9ddc532e7a3f7b3f234747f2e01ba136861c))

# [0.14.0-beta.6](https://github.com/scc-digitalhub/digitalhub-cli/compare/0.14.0-beta.5...0.14.0-beta.6) (2025-10-22)

### Bug Fixes

* **viper:** add new field ini_source. This prevent to update environment when ini file is created from envs ([bfa13f7](https://github.com/scc-digitalhub/digitalhub-cli/commit/bfa13f794ada8eb0db347566d87684326a37056f))
* **viper:** when ini file not present read directly from env without taking into account  .wellknown configuration ([12c7a5b](https://github.com/scc-digitalhub/digitalhub-cli/commit/12c7a5bb033a09222436dbeed3630dba5988476e))

### Features

* **viper:** give precedence to env over ini file ([8522a10](https://github.com/scc-digitalhub/digitalhub-cli/commit/8522a10e408934bbc673abc64c146794f3457a4d))

# [0.14.0-beta.5](https://github.com/scc-digitalhub/digitalhub-cli/compare/0.14.0-beta.4...0.14.0-beta.5) (2025-10-20)

### Bug Fixes

* Add missed file globalprogress ([6975d23](https://github.com/scc-digitalhub/digitalhub-cli/commit/6975d234395e1ec0e505aa2e3faba039b7e0d0e4))
* command help descriptions ([ab817cd](https://github.com/scc-digitalhub/digitalhub-cli/commit/ab817cd04345d49b35032070f39b7c46fb2d830f))
* Delete unecessary files ([fd23848](https://github.com/scc-digitalhub/digitalhub-cli/commit/fd23848ad0e932c9d90c95926533c6db9caa48d8))

### Features

* add download and upload info with percentage when no --verbose flag is passed to the command ([e1493aa](https://github.com/scc-digitalhub/digitalhub-cli/commit/e1493aacb295d0609ca9b1996a0b9991471998d6))
* **download:** add continuation token ([b290db5](https://github.com/scc-digitalhub/digitalhub-cli/commit/b290db5e0d157532e48a26425a1a8b32312dec19))
* **download:** Add download information when flag -v is set on download command. ([19eacd5](https://github.com/scc-digitalhub/digitalhub-cli/commit/19eacd5d258566e046b310f91ee8835583aa55d0))
* **download:** Align logs ([2cd482b](https://github.com/scc-digitalhub/digitalhub-cli/commit/2cd482bb6a23c884d3903e0487f6909f7cabe65e))
* **download:** fix continuation token download. Strip s3 prefix from local path. ([043cc43](https://github.com/scc-digitalhub/digitalhub-cli/commit/043cc43378238c4bfb90ccd09adbe7d948a3f6d7))
* **upload:** Align upload with -v / --verbose option as download. Add some real time upload status ([985fde4](https://github.com/scc-digitalhub/digitalhub-cli/commit/985fde4525c896a0e2a300ba9d5d206a3c1d0e31))

# [0.14.0-beta.3](https://github.com/scc-digitalhub/digitalhub-cli/compare/0.14.0-beta.2...0.14.0-beta.3) (2025-09-17)

### Features

* **download:** report local target paths for each file (short/json/yaml); preserve save logic and handle S3 directories ([6937730](https://github.com/scc-digitalhub/digitalhub-cli/commit/6937730287d053fa0e0d0ce7b791dd791679bf7e))

# [0.14.0-beta.2](https://github.com/scc-digitalhub/digitalhub-cli/compare/0.14.0-beta.1...0.14.0-beta.2) (2025-09-17)

### Features

* **config:** lazy bootstrap INI from well-known; simplify env resolution ([9204eaf](https://github.com/scc-digitalhub/digitalhub-cli/commit/9204eaf145935df20434817f81f2a3dcc0a7ca53))

# [0.14.0-beta.1](https://github.com/scc-digitalhub/digitalhub-cli/compare/0.13.3...0.14.0-beta.1) (2025-09-16)

### Bug Fixes

* Add user workdir on dockerfile ([fa57908](https://github.com/scc-digitalhub/digitalhub-cli/commit/fa579086463ef85fa0ab19cd185cb950ae950624))
* comment debug function ([dd871f8](https://github.com/scc-digitalhub/digitalhub-cli/commit/dd871f85cd2be8491d9da7ececea5548f2e11679))
* Goreleaser.yaml updated and tested. It produce all the archives for all platform ([c2aba34](https://github.com/scc-digitalhub/digitalhub-cli/commit/c2aba348576eba45ed04d918cd2e1d200461ca1e))
* Update Dockerfile with non root user ([2661b46](https://github.com/scc-digitalhub/digitalhub-cli/commit/2661b4628abd08be3bee7aece44e8a6cd5bff484))
* update goreleaser.yaml file. Try to generate zip and tgz for all arch ([ccb6a7c](https://github.com/scc-digitalhub/digitalhub-cli/commit/ccb6a7c32f0fbd6a8fa0e79bb0b20402b55d4901))
* update gorlease.yaml file to accept beta version ([526616e](https://github.com/scc-digitalhub/digitalhub-cli/commit/526616e91adfaa8b5b28de27847af67dbe43accc))
* WriteIniFromStruct should not store the UpdateEnvKey ([7da58f9](https://github.com/scc-digitalhub/digitalhub-cli/commit/7da58f9c60eb05ecfec795e2a7017f11b001ca29))

### Features

* read config from env if ini missing and map OpenID fields to DHCORE_* keys ([9049480](https://github.com/scc-digitalhub/digitalhub-cli/commit/9049480d2c6b7e09b8a7a817257ba0dcef747f49))
* typed config, allowlisted INI, sane env updates. dhcli work also if no INI configuration has been found ([b73e1cc](https://github.com/scc-digitalhub/digitalhub-cli/commit/b73e1ccc235669de900ab0e7949539ea47108431))

## [0.13.3](https://github.com/scc-digitalhub/digitalhub-cli/compare/0.13.2...0.13.3) (2025-09-09)

## [0.13.2](https://github.com/scc-digitalhub/digitalhub-cli/compare/0.13.1...0.13.2) (2025-09-02)

### Bug Fixes

* Change release.yaml to execute action on push in main branch ([6668126](https://github.com/scc-digitalhub/digitalhub-cli/commit/66681264857295f55fb2cd39851ccd7e3bf735b8))
* Change release.yaml to execute action on push new tags ([7f40af6](https://github.com/scc-digitalhub/digitalhub-cli/commit/7f40af6340f3d8a80911f618fbfd4c526654d0cb))
* Clean code ([03a6ff7](https://github.com/scc-digitalhub/digitalhub-cli/commit/03a6ff7fe5bf7c280868afe88ca22beee4d63640))
* path; clean error ([970eb27](https://github.com/scc-digitalhub/digitalhub-cli/commit/970eb2775e6b2560e39ac85571a6c8ad8db2cd07))
* Update release.yaml file with build process ([ab75edb](https://github.com/scc-digitalhub/digitalhub-cli/commit/ab75edbf577b7d9c37508f7e775651238b391b6d))

### Features

* Modify release.yaml file ([ca6625b](https://github.com/scc-digitalhub/digitalhub-cli/commit/ca6625bbc27c73e18b568b05875d660b1629e689))

## [0.13.1](https://github.com/scc-digitalhub/digitalhub-cli/compare/0.13.0...0.13.1) (2025-08-26)

### Bug Fixes

* hardcoded list of resources ([33223ae](https://github.com/scc-digitalhub/digitalhub-cli/commit/33223ae562d41b75e4783484f0cb79c017f5492e))

# [0.13.0](https://github.com/scc-digitalhub/digitalhub-cli/compare/0.11.1...0.13.0) (2025-08-12)

### Bug Fixes

* align code between main and upload branch ([8b13aac](https://github.com/scc-digitalhub/digitalhub-cli/commit/8b13aac4442e238e01a759dbb3fbf09ba6acd645))
* args for refresh/remove/use, login formatting, descriptions ([b205fb9](https://github.com/scc-digitalhub/digitalhub-cli/commit/b205fb9466c6a0dde264640e621820fdf7ec688b))
* argument checks, descriptions ([b585acc](https://github.com/scc-digitalhub/digitalhub-cli/commit/b585accf539f7b07d6840b3a30c7d6dcc2d8d8a1))
* call ReflectValue on updateEnvironment function in envupdate. Problem in serializing complex object ([baebea9](https://github.com/scc-digitalhub/digitalhub-cli/commit/baebea998da98a72fde6d5ab03e3ccd105b0223a))
* change get.go file, rename print Path into Key ([a5f516f](https://github.com/scc-digitalhub/digitalhub-cli/commit/a5f516f584fce3813ee524e0a57c015d52a1ff6a))
* clean project ([649fac8](https://github.com/scc-digitalhub/digitalhub-cli/commit/649fac8a9a216eed1aa6e3e3619b1843c5032545))
* delete useless files ([45ec488](https://github.com/scc-digitalhub/digitalhub-cli/commit/45ec488144a27bbd0c3b9de4ff2ed34306c9df33))
* delete useless files ([a85c2f1](https://github.com/scc-digitalhub/digitalhub-cli/commit/a85c2f1dedb66509de347649ff49b03a95bed9c5))
* error when metrics are missing ([f608764](https://github.com/scc-digitalhub/digitalhub-cli/commit/f608764c7cdcc5d65fbcaf90c1ef8c547aa86a60))
* fix download artifact by id ([eae0602](https://github.com/scc-digitalhub/digitalhub-cli/commit/eae06020662c6d9f2a74d8c9f4411fcae9678dca))
* fix login, skip non necessaries key coming from the token ([f8ef6e8](https://github.com/scc-digitalhub/digitalhub-cli/commit/f8ef6e8a994d80c2c6ef294c542d8052efced0f5))
* fix path on artifact upload, path contains also the filename. ([3dad446](https://github.com/scc-digitalhub/digitalhub-cli/commit/3dad446c2bd9e845ac0449b1038b34884426842b))
* Fix upload directory on aws s3, strip out the baseDir from the path ([67cbab4](https://github.com/scc-digitalhub/digitalhub-cli/commit/67cbab4e0e63d3e2a55b24bff074f725f2f58cf7))
* leaner config.json structure ([3c046e8](https://github.com/scc-digitalhub/digitalhub-cli/commit/3c046e86c31e0e114b763c5020fcc1712f88ad5d))
* log, stop, resume as separate commands ([78ce1e6](https://github.com/scc-digitalhub/digitalhub-cli/commit/78ce1e658f625807dbfad882cdc4ecbf3f3909ae))
* metrics, list format, log msg in register/use/remove/list-env ([341a359](https://github.com/scc-digitalhub/digitalhub-cli/commit/341a3592ad908c562d712ccbacabffa56ed72fa7))
* more meaningful error when status != 200 ([f1f4423](https://github.com/scc-digitalhub/digitalhub-cli/commit/f1f4423ecdcb3ebfa9c2e4268aaafc7f329d3094))
* print state in stop/resume, clear console in log -f ([39ba7d3](https://github.com/scc-digitalhub/digitalhub-cli/commit/39ba7d3819babfad5714847c20c1a93985f1c972))
* refresh using viper; enforce projects on log/metrics/resume/stop ([cb30267](https://github.com/scc-digitalhub/digitalhub-cli/commit/cb30267f5105ef2e15807c89249f88a65e7ef107))
* remove test folder ([a02893c](https://github.com/scc-digitalhub/digitalhub-cli/commit/a02893cca185ba74b4f1f4e47cbcb45a25c94231))
* store path on artifact status and fix the path ([b503f73](https://github.com/scc-digitalhub/digitalhub-cli/commit/b503f73a7f41e16b449ad92a77617f49ae745c5f))
* strip out prefix on download. In case of folder the artifact id is not stored anymore as a local path ([fb68f30](https://github.com/scc-digitalhub/digitalhub-cli/commit/fb68f30a3e161f81540961433e6c5db1e87d0af1))
* test download of multiple files ([2991d98](https://github.com/scc-digitalhub/digitalhub-cli/commit/2991d980eb48318d3e5278561ac100f4b28b0b7a))
* use fmt.Sprint instead of ReflectValue (should not be used anymore) ([2d333b7](https://github.com/scc-digitalhub/digitalhub-cli/commit/2d333b7e9dece7bbbbf435d09faf5d09addef4d1))
* yaml output comments ([d1dc943](https://github.com/scc-digitalhub/digitalhub-cli/commit/d1dc9434fc258d7fffce62b73ecb21fe916b0ca0))

### Features

* add downloader utility with s3 and http file download, modify download.go file to discriminate based on the Prefix the source ([941f821](https://github.com/scc-digitalhub/digitalhub-cli/commit/941f821bb3dc42eeab64c2b85aa9f20f83a1af6f))
* add List files folder in S3 ([95adda7](https://github.com/scc-digitalhub/digitalhub-cli/commit/95adda79bbe9f29d072cd139e062993f0eae020a))
* Add mime-type to upload function ([314aa28](https://github.com/scc-digitalhub/digitalhub-cli/commit/314aa28c0dd2212245789b2947a83f20c8812994))
* Added upload feature for artifact, an artifact can be upload on minio if already exist on core or no. In the second case artifact will be also created on core ([32a75dc](https://github.com/scc-digitalhub/digitalhub-cli/commit/32a75dcc9583e804b212acc6919d9b0974002839))
* Configure multipart uploading when file is bigger than 100MB ([7012271](https://github.com/scc-digitalhub/digitalhub-cli/commit/701227125d565c391546e8ae61002fe5ea6faa7d))
* Introduce Viper to load INI configuration and merge it with exported environment variables. Refactor all services: remove LoadIniConfig, modify BuildCoreUrl, and update all functions that previously took section as a parameter. From now on, Viper is used to retrieve the current environment and all related information. The INI file is loaded once in the root command. If --env is passed to use a different environment, the switch logic is handled at the root level. ([c240159](https://github.com/scc-digitalhub/digitalhub-cli/commit/c24015911627263e53fb810ebb69b96da7a88438))
* metrics ([35f7a1b](https://github.com/scc-digitalhub/digitalhub-cli/commit/35f7a1b74ea9d53bf10c8a0f33f2668d45abbb47))
* run operations and logs ([5cf341f](https://github.com/scc-digitalhub/digitalhub-cli/commit/5cf341f9dab067192284d5c49f1ce098f04e16db))
* working on s3 artifact download, working on another library for commands ([e21c76e](https://github.com/scc-digitalhub/digitalhub-cli/commit/e21c76eefab4281679e1bcb87a7e93f9018363c7))
* Working on upload command. Refactor flags ([9e4a765](https://github.com/scc-digitalhub/digitalhub-cli/commit/9e4a765569a3fb7f3f6a3e0e2854363f06c1b599))

## [0.11.1](https://github.com/scc-digitalhub/digitalhub-cli/compare/0.11.0...0.11.1) (2025-06-09)

### Bug Fixes

* update ini if a hour has passed ([48296f9](https://github.com/scc-digitalhub/digitalhub-cli/commit/48296f9dbd32166c1fa0b7926e572aa52952e662))

# [0.11.0](https://github.com/scc-digitalhub/digitalhub-cli/compare/0.10.3...0.11.0) (2025-06-03)

### Bug Fixes

* cascade warning, restrict resources to list ([7558ef0](https://github.com/scc-digitalhub/digitalhub-cli/commit/7558ef0a9d29263ae22982804b71892fde27933d))
* create + update ([b8e2524](https://github.com/scc-digitalhub/digitalhub-cli/commit/b8e25246ae183868cfed4e1c4edec0883a059940))
* debug messages in stderr ([b168ad0](https://github.com/scc-digitalhub/digitalhub-cli/commit/b168ad08a2d5bc68e258ad22135a316bd7286143))
* delete endpoints, docs ([06236d2](https://github.com/scc-digitalhub/digitalhub-cli/commit/06236d292df7fcf1d47c812538f679525202f341))
* environment name missing ([51e84f0](https://github.com/scc-digitalhub/digitalhub-cli/commit/51e84f09b92893a3dac2bafdba78067eaee59695))
* error when dhcore_version has unexpected format ([901e0c5](https://github.com/scc-digitalhub/digitalhub-cli/commit/901e0c57bd9ba520da054c32563e0ed41c1a2b6f))
* global ini config ([eb99f76](https://github.com/scc-digitalhub/digitalhub-cli/commit/eb99f76fc02a7b05dc58d8f37ffc3436025f8563))
* handle versions in list/get/delete; create projects by name; cleanup ([5a8f6ee](https://github.com/scc-digitalhub/digitalhub-cli/commit/5a8f6eed0faac0d0dc48cb9f51213a2f840d7430))
* output via print + newline ([10acc22](https://github.com/scc-digitalhub/digitalhub-cli/commit/10acc229d7bf0a307847cd25af38bbd87826a345))
* register arrays as comma-sep strings ([ed46af4](https://github.com/scc-digitalhub/digitalhub-cli/commit/ed46af454cc51c4cba5f77fb81c8ff4c8680a976))
* register saves ts ([fba987d](https://github.com/scc-digitalhub/digitalhub-cli/commit/fba987df0bac4b62f632e50ead12b5c69c2d682e))
* short header case, empty lines, yaml comment ([88f06ca](https://github.com/scc-digitalhub/digitalhub-cli/commit/88f06ca42d78ab8528b026a8351f6d35d5aa419a))
* sort by updated/asc when short ([d31bf8e](https://github.com/scc-digitalhub/digitalhub-cli/commit/d31bf8e70d9f672f0a4654f962b00d4d8b98925d))
* usage ([ea66c9e](https://github.com/scc-digitalhub/digitalhub-cli/commit/ea66c9e86d2ac0ff66c32902c147f4c3af1af42f))
* version comparison ([e594043](https://github.com/scc-digitalhub/digitalhub-cli/commit/e5940436aa7593b112c3561a4b7725ce949309fe))

### Features

* crud operations ([23190da](https://github.com/scc-digitalhub/digitalhub-cli/commit/23190da7a00ee60773e971346602fd179b609af7))
* get + list ([0570d97](https://github.com/scc-digitalhub/digitalhub-cli/commit/0570d975c0c9ba09c37328b09560f109f5d19cf0))

## [0.10.3](https://github.com/scc-digitalhub/digitalhub-cli/compare/0.10.2...0.10.3) (2025-04-15)

### Bug Fixes

* login for windows ([1eba089](https://github.com/scc-digitalhub/digitalhub-cli/commit/1eba089e652a374f34d5ba3b501aa394881c5670))

## [0.10.2](https://github.com/scc-digitalhub/digitalhub-cli/compare/0.10.1...0.10.2) (2025-04-15)

### Bug Fixes

* register re-initializes a section ([23d3a6a](https://github.com/scc-digitalhub/digitalhub-cli/commit/23d3a6a00307ae62fa600ba6975c2eac63d3a2ca))
* values updated on register/login ([f964fcf](https://github.com/scc-digitalhub/digitalhub-cli/commit/f964fcf2ad446bb482b8b5c5a79a90cd2985760b))

## [0.10.1](https://github.com/scc-digitalhub/digitalhub-cli/compare/0.10.0...0.10.1) (2025-04-02)

### Bug Fixes

* login asks confirmation before opening page in browser ([fb7ff48](https://github.com/scc-digitalhub/digitalhub-cli/commit/fb7ff48c3530789aba1f8d03b7b7d857006d290d))

# [0.10.0](https://github.com/scc-digitalhub/digitalhub-cli/compare/0.10.0-beta-gr...0.10.0) (2025-04-01)

### Bug Fixes

* init ([f9d4f2c](https://github.com/scc-digitalhub/digitalhub-cli/commit/f9d4f2cc04d8860beca214eed3669e181a1a7bc9))
* init packages ([94616bc](https://github.com/scc-digitalhub/digitalhub-cli/commit/94616bcff5b57f5009ad1a00d56be05aff44d1f8))
* removed debug message ([580849f](https://github.com/scc-digitalhub/digitalhub-cli/commit/580849fa4cfdc7e56e740aa30d1c4738e581532c))
* scope, log, -h, remove msg ([c493014](https://github.com/scc-digitalhub/digitalhub-cli/commit/c49301450efb0e94362979faaa2991ca6f795d93))
* style for login success ([c9e39e0](https://github.com/scc-digitalhub/digitalhub-cli/commit/c9e39e066de7d581764c7c3edf5bddd8b63e1562))

### Features

* list-env ([d175e53](https://github.com/scc-digitalhub/digitalhub-cli/commit/d175e53a57ec14ac7c2d5a743380dae32075719f))

# [0.10.0-beta-gr](https://github.com/scc-digitalhub/digitalhub-cli/compare/d2b6c14ce3ad22775ad33fcece696c0797072a6e...0.10.0-beta-gr) (2025-03-25)

### Bug Fixes

* added scope to register and handling slices ([d938f8a](https://github.com/scc-digitalhub/digitalhub-cli/commit/d938f8a9e9b5afac2345508651b528dbd37d19d8))
* change client id with hardcoded client ([17da72c](https://github.com/scc-digitalhub/digitalhub-cli/commit/17da72cc7b7e7e8389ebf866bdc48b5f81066d05))
* init checks versions and asks confirmation; ini added to gitignore automatically ([755d862](https://github.com/scc-digitalhub/digitalhub-cli/commit/755d862c28573c50907bfa0b274b53b1de977a96))
* init now takes version from environment; register handles response errors ([15ae071](https://github.com/scc-digitalhub/digitalhub-cli/commit/15ae071db3205b65f869672c32050794d677f38a))
* init: runtimes + --pre flag ([6cff8c0](https://github.com/scc-digitalhub/digitalhub-cli/commit/6cff8c0efac08bc79ccadc326f9f09316d0a14e9))
* initialize ini file on register if missing ([c6e02ea](https://github.com/scc-digitalhub/digitalhub-cli/commit/c6e02ea56a84e1d9633806953421271958840551))
* prefixes and log message when successful ([f38a6e6](https://github.com/scc-digitalhub/digitalhub-cli/commit/f38a6e695139b58fb43eafcec2c77611420511d1))
* removed auto-close window ([e1d7857](https://github.com/scc-digitalhub/digitalhub-cli/commit/e1d78570abcd6d9857263e1cb52ec1f2d61f55e5))

### Features

* Create CLI for digitalhub ([d2b6c14](https://github.com/scc-digitalhub/digitalhub-cli/commit/d2b6c14ce3ad22775ad33fcece696c0797072a6e))
* init ([ba4f279](https://github.com/scc-digitalhub/digitalhub-cli/commit/ba4f279d687e54b347e25fe7b03053cdd5843492))
* refresh ([8481599](https://github.com/scc-digitalhub/digitalhub-cli/commit/8481599f7957c9922d7c9488197d01501f8ed3cd))
* register in ini, logout ([2472e6f](https://github.com/scc-digitalhub/digitalhub-cli/commit/2472e6f060924659f83ed93f379db864027f3a36))
* remove and use commands, default environment ([6a2889c](https://github.com/scc-digitalhub/digitalhub-cli/commit/6a2889c58a7ee2d75889abf27b943a36e7c10ca3))
* support core as provider ([71fecfb](https://github.com/scc-digitalhub/digitalhub-cli/commit/71fecfbdca1cb52d5648789f192b3ed2cdcbdc48))
