# Changelog

## [1.6.1](https://github.com/vijayet1/az-pim-cli/compare/v1.13.0...v1.6.1) (2026-04-14)


### ⚠ BREAKING CHANGES

* use proper terms for 'azure resources' type ([#59](https://github.com/vijayet1/az-pim-cli/issues/59))

### Features

* activate roles ([7cdb3be](https://github.com/vijayet1/az-pim-cli/commit/7cdb3be77fe393028096d066192a6c1631b3ac3d))
* add 'version' command ([#30](https://github.com/vijayet1/az-pim-cli/issues/30)) ([e24a15f](https://github.com/vijayet1/az-pim-cli/commit/e24a15f6fb1aa020e6e7191080c3b56363eac355))
* add reason to activate command ([#4](https://github.com/vijayet1/az-pim-cli/issues/4)) ([8b43135](https://github.com/vijayet1/az-pim-cli/commit/8b4313595e4b534c304619c973d42e2c8e8b1d35))
* add support for additional Azure environments (us gov, china) ([#111](https://github.com/vijayet1/az-pim-cli/issues/111)) ([b180576](https://github.com/vijayet1/az-pim-cli/commit/b1805761e280d6c0db39cd30a51d2b35cd0e7685))
* add support for setting start-time ([#81](https://github.com/vijayet1/az-pim-cli/issues/81)) ([ee8a4a9](https://github.com/vijayet1/az-pim-cli/commit/ee8a4a914be91c7ef3e7d84da3cdcd66b8e31fe9))
* add support for specifying 'ticket number' and 'ticket system' ([#56](https://github.com/vijayet1/az-pim-cli/issues/56)) ([a62c52f](https://github.com/vijayet1/az-pim-cli/commit/a62c52ff158a018d46598fa6c631ebc020c52d53))
* **auth:** get token for pim governance roles ([#113](https://github.com/vijayet1/az-pim-cli/issues/113)) ([ae74121](https://github.com/vijayet1/az-pim-cli/commit/ae74121bed242016242dd13612c46f56e6bcc672))
* check for various request status types ([#14](https://github.com/vijayet1/az-pim-cli/issues/14)) ([57e4472](https://github.com/vijayet1/az-pim-cli/commit/57e447247280dc092cc2b9ee817a53b599b47ae9))
* dry-run for 'activate' ([#22](https://github.com/vijayet1/az-pim-cli/issues/22)) ([05c4095](https://github.com/vijayet1/az-pim-cli/commit/05c40956017909a14f3015f2de10c4a5e43303e2))
* **goreleaser:** include binaries in release ([#101](https://github.com/vijayet1/az-pim-cli/issues/101)) ([808cdf1](https://github.com/vijayet1/az-pim-cli/commit/808cdf12ddd1b95f2e96b83b325683190d05f68f))
* improved error messages and logging ([#68](https://github.com/vijayet1/az-pim-cli/issues/68)) ([bbeea03](https://github.com/vijayet1/az-pim-cli/commit/bbeea03b138d28653cc667954cd56cc25a9d9fa5))
* list eligible roles ([eb3e15a](https://github.com/vijayet1/az-pim-cli/commit/eb3e15ae475d065613c1cb816dc6082e9d008c76))
* Support for Entra roles ([#61](https://github.com/vijayet1/az-pim-cli/issues/61)) ([dd9ed19](https://github.com/vijayet1/az-pim-cli/commit/dd9ed193c7bee3a85ad3cc62ada4bc2630378393))
* support for PIM Entra groups ([#16](https://github.com/vijayet1/az-pim-cli/issues/16)) ([6fddc87](https://github.com/vijayet1/az-pim-cli/commit/6fddc870a990bc6065b8dd053544fc141421428f))
* support new Azure Entra ID PIM API ([#6](https://github.com/vijayet1/az-pim-cli/issues/6)) ([700323c](https://github.com/vijayet1/az-pim-cli/commit/700323cc0c90674f8d1b8fd9db6db96933e15bbc))
* use az-cli for auth ([95e7553](https://github.com/vijayet1/az-pim-cli/commit/95e7553cd7142b0ba35f7054f4762b23764804d3))
* use proper terms for 'azure resources' type ([#59](https://github.com/vijayet1/az-pim-cli/issues/59)) ([6411902](https://github.com/vijayet1/az-pim-cli/commit/641190289f99d2599d7dd789c5c3ea10845746ae))


### Bug Fixes

* **activate:** Role selection on `activate` selects incorrect role ([#8](https://github.com/vijayet1/az-pim-cli/issues/8)) ([6cb1079](https://github.com/vijayet1/az-pim-cli/commit/6cb1079b62cabf219232c9e829198d70b4b122e8))
* **build/goreleaser:** remove incorrect template key usage ([#103](https://github.com/vijayet1/az-pim-cli/issues/103)) ([ef49cb2](https://github.com/vijayet1/az-pim-cli/commit/ef49cb2a81b455215559202298b126e99d97182c))
* fix casing role on activate ([#3](https://github.com/vijayet1/az-pim-cli/issues/3)) ([9d92cff](https://github.com/vijayet1/az-pim-cli/commit/9d92cff54a4515eb44e6226c623fe8f59cf9817c))
* **pim-client:** resolve invalid logic for building a request ([#76](https://github.com/vijayet1/az-pim-cli/issues/76)) ([ece6a96](https://github.com/vijayet1/az-pim-cli/commit/ece6a96be07f771ce9308f47750ff41c2c4676d8))
* **pim-client:** resolve invalid logic for building a request dynamically ([ece6a96](https://github.com/vijayet1/az-pim-cli/commit/ece6a96be07f771ce9308f47750ff41c2c4676d8))
* prevent panic from JWT parsing ([01a09ca](https://github.com/vijayet1/az-pim-cli/commit/01a09ca5aa46e437136f264f9bc412a6dc34a86b))
* prevent panic from JWT parsing ([#96](https://github.com/vijayet1/az-pim-cli/issues/96)) ([01a09ca](https://github.com/vijayet1/az-pim-cli/commit/01a09ca5aa46e437136f264f9bc412a6dc34a86b))
* trailing whitespace issue ([adb9d8e](https://github.com/vijayet1/az-pim-cli/commit/adb9d8ec0e80252bd62cd955b48fe8504862fc1f))
* use exact matching for the role selection ([#12](https://github.com/vijayet1/az-pim-cli/issues/12)) ([0bf37e6](https://github.com/vijayet1/az-pim-cli/commit/0bf37e6db2e648179442326c0b101328e4fd7e82))
* use right scope to generate pim token for groups and roles ([#115](https://github.com/vijayet1/az-pim-cli/issues/115)) ([240c342](https://github.com/vijayet1/az-pim-cli/commit/240c342e0a8b512c3e8b39ebc8ed212b7d5c0b28))


### Documentation

* **github:** add project guidelines ([#31](https://github.com/vijayet1/az-pim-cli/issues/31)) ([5fc195b](https://github.com/vijayet1/az-pim-cli/commit/5fc195bda5e78fd66b0fc996b3259d380b40f102))
* initial docs ([c315b5c](https://github.com/vijayet1/az-pim-cli/commit/c315b5c44dab5102e8a7678c09e3c81d35f87a09))


### Code Refactoring

* create interface for azure client ([#72](https://github.com/vijayet1/az-pim-cli/issues/72)) ([7391369](https://github.com/vijayet1/az-pim-cli/commit/7391369453d3d24dd17e024e48100260d68da4da))


### Chores

* add FUNDING file ([#74](https://github.com/vijayet1/az-pim-cli/issues/74)) ([7462bc6](https://github.com/vijayet1/az-pim-cli/commit/7462bc68342c3a73cbd399c54f6efc6e54491538))
* added LICENSE ([#25](https://github.com/vijayet1/az-pim-cli/issues/25)) ([7060302](https://github.com/vijayet1/az-pim-cli/commit/70603021f5a49e9e424bbf68321331102bcdb4e6))
* bootstrap releases for path: . ([#34](https://github.com/vijayet1/az-pim-cli/issues/34)) ([06e230a](https://github.com/vijayet1/az-pim-cli/commit/06e230a76a9a490be5a4fbf9a88a9762f0fe6a36))
* create release v1.6.1 ([#89](https://github.com/vijayet1/az-pim-cli/issues/89)) ([4b0316c](https://github.com/vijayet1/az-pim-cli/commit/4b0316cbbfb5091b9fb301bb901a39c1bfd58d91))
* **deps:** bump golang.org/x/crypto from 0.24.0 to 0.31.0 in the go_modules group ([#79](https://github.com/vijayet1/az-pim-cli/issues/79)) ([3c64593](https://github.com/vijayet1/az-pim-cli/commit/3c64593cd8ace5b90a78b9dbbcdd164e3388f91c))
* **deps:** bump golang.org/x/crypto in the go_modules group ([3c64593](https://github.com/vijayet1/az-pim-cli/commit/3c64593cd8ace5b90a78b9dbbcdd164e3388f91c))
* **deps:** bump golang.org/x/net from 0.22.0 to 0.23.0 ([#11](https://github.com/vijayet1/az-pim-cli/issues/11)) ([26703f1](https://github.com/vijayet1/az-pim-cli/commit/26703f15f7c11b1d2296ef393a1efa52330a29cd))
* **deps:** bump golang.org/x/net from 0.26.0 to 0.33.0 in the go_modules group ([#84](https://github.com/vijayet1/az-pim-cli/issues/84)) ([9432118](https://github.com/vijayet1/az-pim-cli/commit/94321184348d8ecf39ffbc7d1738eeb56c27f27b))
* **deps:** bump golang.org/x/net in the go_modules group ([9432118](https://github.com/vijayet1/az-pim-cli/commit/94321184348d8ecf39ffbc7d1738eeb56c27f27b))
* **deps:** upgrade dependencies ([#99](https://github.com/vijayet1/az-pim-cli/issues/99)) ([f4cf17c](https://github.com/vijayet1/az-pim-cli/commit/f4cf17c62213f3273b2c9760cec9cffcffb39a9a))
* **docs:** cleanup ([4fc7ecd](https://github.com/vijayet1/az-pim-cli/commit/4fc7ecdfe6c7f4e9f94980908f6437234cf0e48f))
* ensure consistent flag names ([#18](https://github.com/vijayet1/az-pim-cli/issues/18)) ([57782e6](https://github.com/vijayet1/az-pim-cli/commit/57782e6c06f119787b04a885bc15eab95224d7c0))
* ignores ([40ce8f7](https://github.com/vijayet1/az-pim-cli/commit/40ce8f7bbb9c7d273e97cabcd10cc6c04c00ade6))
* **main:** release 1.0.1 ([#49](https://github.com/vijayet1/az-pim-cli/issues/49)) ([e6d96fc](https://github.com/vijayet1/az-pim-cli/commit/e6d96fc864f0b8cb5bad5d53cceb6fa094937cd7))
* **main:** release 1.1.0 ([#57](https://github.com/vijayet1/az-pim-cli/issues/57)) ([04a6e49](https://github.com/vijayet1/az-pim-cli/commit/04a6e495f6f1698e36f8941784aee5a6161d29cc))
* **main:** release 1.10.0 ([#108](https://github.com/vijayet1/az-pim-cli/issues/108)) ([776315c](https://github.com/vijayet1/az-pim-cli/commit/776315caaeff1bfeb72203c0e14307e03e1ee823))
* **main:** release 1.11.0 ([#112](https://github.com/vijayet1/az-pim-cli/issues/112)) ([2e9b90b](https://github.com/vijayet1/az-pim-cli/commit/2e9b90b265f43c9102060cfed77fd779ad06e14c))
* **main:** release 1.12.0 ([#114](https://github.com/vijayet1/az-pim-cli/issues/114)) ([acb5393](https://github.com/vijayet1/az-pim-cli/commit/acb53937713f11b85ca90fdecb227a104fec1b5a))
* **main:** release 1.13.0 ([#117](https://github.com/vijayet1/az-pim-cli/issues/117)) ([18a1dbf](https://github.com/vijayet1/az-pim-cli/commit/18a1dbffc011f734700e193bac75b8e695aab2b3))
* **main:** release 1.2.0 ([#63](https://github.com/vijayet1/az-pim-cli/issues/63)) ([dd15832](https://github.com/vijayet1/az-pim-cli/commit/dd158323226f64870c200150a746bc710cfce2be))
* **main:** release 1.3.0 ([#65](https://github.com/vijayet1/az-pim-cli/issues/65)) ([ec7738f](https://github.com/vijayet1/az-pim-cli/commit/ec7738f7f35898ed35c47c7f24e643e45042bc80))
* **main:** release 1.4.0 ([#70](https://github.com/vijayet1/az-pim-cli/issues/70)) ([f91ceed](https://github.com/vijayet1/az-pim-cli/commit/f91ceed3c10438692de46dd53818cf21989f734f))
* **main:** release 1.5.0 ([#77](https://github.com/vijayet1/az-pim-cli/issues/77)) ([a71add4](https://github.com/vijayet1/az-pim-cli/commit/a71add4ff6ff97f2195af980f89a181966ebeed5))
* **main:** release 1.6.0 ([#85](https://github.com/vijayet1/az-pim-cli/issues/85)) ([69fb797](https://github.com/vijayet1/az-pim-cli/commit/69fb79750cecfc18445bd82fc357064f1434afe5))
* **main:** release 1.6.1 ([#90](https://github.com/vijayet1/az-pim-cli/issues/90)) ([e2d9fb9](https://github.com/vijayet1/az-pim-cli/commit/e2d9fb92fb6764283fb8752348a2af48ed2de311))
* **main:** release 1.7.0 ([#92](https://github.com/vijayet1/az-pim-cli/issues/92)) ([5019657](https://github.com/vijayet1/az-pim-cli/commit/5019657fa2a12615027a3944d722a87877571f05))
* **main:** release 1.8.0 ([#100](https://github.com/vijayet1/az-pim-cli/issues/100)) ([4aac9fc](https://github.com/vijayet1/az-pim-cli/commit/4aac9fc19b405b0cb5ad36a8c3bc8bade05352ca))
* **main:** release 1.9.0 ([#106](https://github.com/vijayet1/az-pim-cli/issues/106)) ([c0889fc](https://github.com/vijayet1/az-pim-cli/commit/c0889fc921553527cd59998915482d4f1eed1f27))
* replace deprecated ioutil package (SA1019) ([b3531c1](https://github.com/vijayet1/az-pim-cli/commit/b3531c1f7f4edcb9304f2a0defd0681f0dbbbf7e))
* update dependencies ([#10](https://github.com/vijayet1/az-pim-cli/issues/10)) ([458c942](https://github.com/vijayet1/az-pim-cli/commit/458c94259d876f2a8114a4582abb2faaecea7913))
* update dependencies ([#28](https://github.com/vijayet1/az-pim-cli/issues/28)) ([a1182f7](https://github.com/vijayet1/az-pim-cli/commit/a1182f7576f841b1b8476adaf22813957d8ffe12))
* upgrade dependencies ([f4cf17c](https://github.com/vijayet1/az-pim-cli/commit/f4cf17c62213f3273b2c9760cec9cffcffb39a9a))
* upgrade dependencies ([#88](https://github.com/vijayet1/az-pim-cli/issues/88)) ([b70f388](https://github.com/vijayet1/az-pim-cli/commit/b70f38815df5cfaaa4093a3c35440131376b0ecf))


### Tests

* add unit tests for the client and utils ([#73](https://github.com/vijayet1/az-pim-cli/issues/73)) ([3768852](https://github.com/vijayet1/az-pim-cli/commit/3768852e616d88803553b4b9b7d1b14f0243df37))


### Linting

* resolve 'empty branch' linter error (SA9003) ([21e9888](https://github.com/vijayet1/az-pim-cli/commit/21e9888fe4f115dc741b553c4d49d8be517b50a2))
* resolve 'unused' linter error ([bce3ccc](https://github.com/vijayet1/az-pim-cli/commit/bce3ccc0698fc74b53185bdbbdc11d239f26b1b5))


### Build

* change versioning strategy ([#64](https://github.com/vijayet1/az-pim-cli/issues/64)) ([f7caa1e](https://github.com/vijayet1/az-pim-cli/commit/f7caa1e6a52aa1619ba7ba41c19ba20b683a77d7))
* fix broken goreleaser builds ([#105](https://github.com/vijayet1/az-pim-cli/issues/105)) ([46cb5b5](https://github.com/vijayet1/az-pim-cli/commit/46cb5b526787e3d299fbd08dcbb76dea0f03bcd6))
* **goreleaser:** include macos-universal binary in release archives ([#107](https://github.com/vijayet1/az-pim-cli/issues/107)) ([6e63b46](https://github.com/vijayet1/az-pim-cli/commit/6e63b4636f28c8d99cf36d61a1d759fa98ac5021))


### Continuous Integration

* add golangci-lint config ([dee68c5](https://github.com/vijayet1/az-pim-cli/commit/dee68c5c4cbf4df8c7e75edf5a7b3813c0dcd457))
* add release-please workflow ([#36](https://github.com/vijayet1/az-pim-cli/issues/36)) ([67f357d](https://github.com/vijayet1/az-pim-cli/commit/67f357d1dfb1a2bc981ad257085757e59d934b90))
* add workflow triggered by merge to main ([#37](https://github.com/vijayet1/az-pim-cli/issues/37)) ([4b24cf9](https://github.com/vijayet1/az-pim-cli/commit/4b24cf90b8a58a5a71c36347149418b233fa038b))
* added semgrep workflow ([#29](https://github.com/vijayet1/az-pim-cli/issues/29)) ([d02c5a1](https://github.com/vijayet1/az-pim-cli/commit/d02c5a153f89d3b3a3862435bf945e145706461f))
* added snyk workflow ([#26](https://github.com/vijayet1/az-pim-cli/issues/26)) ([0c841f0](https://github.com/vijayet1/az-pim-cli/commit/0c841f05a8d6c18919af65c83ca859708df24f5c))
* config for goreleaser ([a86d487](https://github.com/vijayet1/az-pim-cli/commit/a86d48760f8465bb83529808525c50413a213c45))
* **fix:** set permissions for scheduled workflow ([#33](https://github.com/vijayet1/az-pim-cli/issues/33)) ([dd16f89](https://github.com/vijayet1/az-pim-cli/commit/dd16f8935c800966b9a2856ab7feae1487ed302e))
* **pre-commit:** add golangci-lint 'full' hook ([2f0ff2b](https://github.com/vijayet1/az-pim-cli/commit/2f0ff2b5efabe84b66170fe18e81d08b3ebe720a))
* **pre-commit:** added conventional-pre-commit hook ([197274f](https://github.com/vijayet1/az-pim-cli/commit/197274ff1f94e2ac5a44084e24c2568a493d5e20))
* **pre-commit:** update hooks ([374490f](https://github.com/vijayet1/az-pim-cli/commit/374490fb0e24b11cb8055dce88077e687ba6ad89))
* **semgrep:** force Node.js 24 runtime for actions to suppress deprecation warning ([7c7bbd4](https://github.com/vijayet1/az-pim-cli/commit/7c7bbd41bd1aba59ef2fe48b55f610a5157ba702))
* set permissions for workflows ([#91](https://github.com/vijayet1/az-pim-cli/issues/91)) ([e4bb4d7](https://github.com/vijayet1/az-pim-cli/commit/e4bb4d7617a0561ae2fad3fb00c1e12d1548d5fc))
* **test:** force Node.js 24 runtime for actions to suppress deprecation warning ([ffaa80d](https://github.com/vijayet1/az-pim-cli/commit/ffaa80d9ba04874efdab7bc76f849590176b9bd6))
* upgrade codeql-action/upload-sarif from v3 to v4 ([3031bdb](https://github.com/vijayet1/az-pim-cli/commit/3031bdbb46a27d356979b7f15c296df74b796265))
* upgrade codeql-action/upload-sarif from v3 to v4 ([4e9e463](https://github.com/vijayet1/az-pim-cli/commit/4e9e463ff63e19589b09efaf7e0c16835c32295f))
* workflow to check for conventional-commits ([96306d9](https://github.com/vijayet1/az-pim-cli/commit/96306d94abbf2c13c7328a9d0cb223f70a838911))
* workflow to run on new tags (release) ([44f25dc](https://github.com/vijayet1/az-pim-cli/commit/44f25dc9647fd5c98d18cd8e3e9793a60a9de86e))
* workflow to run on pull requests ([54c48a4](https://github.com/vijayet1/az-pim-cli/commit/54c48a45d18e0fd0b3e4d5c6eaecd665bb640384))
* workflow to run pre-commit ([54d380b](https://github.com/vijayet1/az-pim-cli/commit/54d380bffb4b5e71bf41b7bca8a14e597dc7bb1f))

## [1.13.0](https://github.com/netr0m/az-pim-cli/compare/v1.12.0...v1.13.0) (2026-04-05)


### Bug Fixes

* use right scope to generate pim token for groups and roles ([#115](https://github.com/netr0m/az-pim-cli/issues/115)) ([240c342](https://github.com/netr0m/az-pim-cli/commit/240c342e0a8b512c3e8b39ebc8ed212b7d5c0b28))

## [1.12.0](https://github.com/netr0m/az-pim-cli/compare/v1.11.0...v1.12.0) (2026-03-27)


### Features

* **auth:** get token for pim governance roles ([#113](https://github.com/netr0m/az-pim-cli/issues/113)) ([ae74121](https://github.com/netr0m/az-pim-cli/commit/ae74121bed242016242dd13612c46f56e6bcc672))

## [1.11.0](https://github.com/netr0m/az-pim-cli/compare/v1.10.0...v1.11.0) (2026-01-01)


### Features

* add support for additional Azure environments (us gov, china) ([#111](https://github.com/netr0m/az-pim-cli/issues/111)) ([b180576](https://github.com/netr0m/az-pim-cli/commit/b1805761e280d6c0db39cd30a51d2b35cd0e7685))

## [1.10.0](https://github.com/netr0m/az-pim-cli/compare/v1.9.0...v1.10.0) (2025-10-03)


### Build

* **goreleaser:** include macos-universal binary in release archives ([#107](https://github.com/netr0m/az-pim-cli/issues/107)) ([6e63b46](https://github.com/netr0m/az-pim-cli/commit/6e63b4636f28c8d99cf36d61a1d759fa98ac5021))

## [1.9.0](https://github.com/netr0m/az-pim-cli/compare/v1.8.0...v1.9.0) (2025-10-03)


### Features

* **goreleaser:** include binaries in release ([#101](https://github.com/netr0m/az-pim-cli/issues/101)) ([808cdf1](https://github.com/netr0m/az-pim-cli/commit/808cdf12ddd1b95f2e96b83b325683190d05f68f))


### Bug Fixes

* **build/goreleaser:** remove incorrect template key usage ([#103](https://github.com/netr0m/az-pim-cli/issues/103)) ([ef49cb2](https://github.com/netr0m/az-pim-cli/commit/ef49cb2a81b455215559202298b126e99d97182c))


### Build

* fix broken goreleaser builds ([#105](https://github.com/netr0m/az-pim-cli/issues/105)) ([46cb5b5](https://github.com/netr0m/az-pim-cli/commit/46cb5b526787e3d299fbd08dcbb76dea0f03bcd6))

## [1.8.0](https://github.com/netr0m/az-pim-cli/compare/v1.7.0...v1.8.0) (2025-10-03)


### Chores

* **deps:** upgrade dependencies ([#99](https://github.com/netr0m/az-pim-cli/issues/99)) ([f4cf17c](https://github.com/netr0m/az-pim-cli/commit/f4cf17c62213f3273b2c9760cec9cffcffb39a9a))
* upgrade dependencies ([f4cf17c](https://github.com/netr0m/az-pim-cli/commit/f4cf17c62213f3273b2c9760cec9cffcffb39a9a))

## [1.7.0](https://github.com/netr0m/az-pim-cli/compare/v1.6.1...v1.7.0) (2025-07-27)


### Bug Fixes

* prevent panic from JWT parsing ([01a09ca](https://github.com/netr0m/az-pim-cli/commit/01a09ca5aa46e437136f264f9bc412a6dc34a86b))
* prevent panic from JWT parsing ([#96](https://github.com/netr0m/az-pim-cli/issues/96)) ([01a09ca](https://github.com/netr0m/az-pim-cli/commit/01a09ca5aa46e437136f264f9bc412a6dc34a86b))


### Continuous Integration

* set permissions for workflows ([#91](https://github.com/netr0m/az-pim-cli/issues/91)) ([e4bb4d7](https://github.com/netr0m/az-pim-cli/commit/e4bb4d7617a0561ae2fad3fb00c1e12d1548d5fc))

## [1.6.1](https://github.com/netr0m/az-pim-cli/compare/v1.6.0...v1.6.1) (2025-05-30)


### Chores

* create release v1.6.1 ([#89](https://github.com/netr0m/az-pim-cli/issues/89)) ([4b0316c](https://github.com/netr0m/az-pim-cli/commit/4b0316cbbfb5091b9fb301bb901a39c1bfd58d91))
* upgrade dependencies ([#88](https://github.com/netr0m/az-pim-cli/issues/88)) ([b70f388](https://github.com/netr0m/az-pim-cli/commit/b70f38815df5cfaaa4093a3c35440131376b0ecf))

## [1.6.0](https://github.com/netr0m/az-pim-cli/compare/v1.5.0...v1.6.0) (2025-02-21)


### Features

* add support for setting start-time ([#81](https://github.com/netr0m/az-pim-cli/issues/81)) ([ee8a4a9](https://github.com/netr0m/az-pim-cli/commit/ee8a4a914be91c7ef3e7d84da3cdcd66b8e31fe9))

## [1.5.0](https://github.com/netr0m/az-pim-cli/compare/v1.4.0...v1.5.0) (2024-11-21)


### Bug Fixes

* **pim-client:** resolve invalid logic for building a request ([#76](https://github.com/netr0m/az-pim-cli/issues/76)) ([ece6a96](https://github.com/netr0m/az-pim-cli/commit/ece6a96be07f771ce9308f47750ff41c2c4676d8))
* **pim-client:** resolve invalid logic for building a request dynamically ([ece6a96](https://github.com/netr0m/az-pim-cli/commit/ece6a96be07f771ce9308f47750ff41c2c4676d8))

## [1.4.0](https://github.com/netr0m/az-pim-cli/compare/v1.3.0...v1.4.0) (2024-11-05)


### Features

* improved error messages and logging ([#68](https://github.com/netr0m/az-pim-cli/issues/68)) ([bbeea03](https://github.com/netr0m/az-pim-cli/commit/bbeea03b138d28653cc667954cd56cc25a9d9fa5))


### Code Refactoring

* create interface for azure client ([#72](https://github.com/netr0m/az-pim-cli/issues/72)) ([7391369](https://github.com/netr0m/az-pim-cli/commit/7391369453d3d24dd17e024e48100260d68da4da))

## [1.3.0](https://github.com/netr0m/az-pim-cli/compare/v1.2.0...v1.3.0) (2024-10-21)


### Features

* Support for Entra roles ([#61](https://github.com/netr0m/az-pim-cli/issues/61)) ([dd9ed19](https://github.com/netr0m/az-pim-cli/commit/dd9ed193c7bee3a85ad3cc62ada4bc2630378393))

## [1.2.0](https://github.com/netr0m/az-pim-cli/compare/v1.1.0...v1.2.0) (2024-10-21)


### ⚠ BREAKING CHANGES

* use proper terms for 'azure resources' type ([#59](https://github.com/netr0m/az-pim-cli/issues/59))

### Features

* use proper terms for 'azure resources' type ([#59](https://github.com/netr0m/az-pim-cli/issues/59)) ([6411902](https://github.com/netr0m/az-pim-cli/commit/641190289f99d2599d7dd789c5c3ea10845746ae))

## [1.1.0](https://github.com/netr0m/az-pim-cli/compare/v1.0.1...v1.1.0) (2024-09-13)


### Features

* add support for specifying 'ticket number' and 'ticket system' ([#56](https://github.com/netr0m/az-pim-cli/issues/56)) ([a62c52f](https://github.com/netr0m/az-pim-cli/commit/a62c52ff158a018d46598fa6c631ebc020c52d53))

## 1.0.1 (2024-07-01)


### Features

* activate roles ([7cdb3be](https://github.com/netr0m/az-pim-cli/commit/7cdb3be77fe393028096d066192a6c1631b3ac3d))
* add 'version' command ([#30](https://github.com/netr0m/az-pim-cli/issues/30)) ([e24a15f](https://github.com/netr0m/az-pim-cli/commit/e24a15f6fb1aa020e6e7191080c3b56363eac355))
* add reason to activate command ([#4](https://github.com/netr0m/az-pim-cli/issues/4)) ([8b43135](https://github.com/netr0m/az-pim-cli/commit/8b4313595e4b534c304619c973d42e2c8e8b1d35))
* check for various request status types ([#14](https://github.com/netr0m/az-pim-cli/issues/14)) ([57e4472](https://github.com/netr0m/az-pim-cli/commit/57e447247280dc092cc2b9ee817a53b599b47ae9))
* dry-run for 'activate' ([#22](https://github.com/netr0m/az-pim-cli/issues/22)) ([05c4095](https://github.com/netr0m/az-pim-cli/commit/05c40956017909a14f3015f2de10c4a5e43303e2))
* list eligible roles ([eb3e15a](https://github.com/netr0m/az-pim-cli/commit/eb3e15ae475d065613c1cb816dc6082e9d008c76))
* support for PIM Entra groups ([#16](https://github.com/netr0m/az-pim-cli/issues/16)) ([6fddc87](https://github.com/netr0m/az-pim-cli/commit/6fddc870a990bc6065b8dd053544fc141421428f))
* support new Azure Entra ID PIM API ([#6](https://github.com/netr0m/az-pim-cli/issues/6)) ([700323c](https://github.com/netr0m/az-pim-cli/commit/700323cc0c90674f8d1b8fd9db6db96933e15bbc))
* use az-cli for auth ([95e7553](https://github.com/netr0m/az-pim-cli/commit/95e7553cd7142b0ba35f7054f4762b23764804d3))


### Bug Fixes

* **activate:** Role selection on `activate` selects incorrect role ([#8](https://github.com/netr0m/az-pim-cli/issues/8)) ([6cb1079](https://github.com/netr0m/az-pim-cli/commit/6cb1079b62cabf219232c9e829198d70b4b122e8))
* fix casing role on activate ([#3](https://github.com/netr0m/az-pim-cli/issues/3)) ([9d92cff](https://github.com/netr0m/az-pim-cli/commit/9d92cff54a4515eb44e6226c623fe8f59cf9817c))
* use exact matching for the role selection ([#12](https://github.com/netr0m/az-pim-cli/issues/12)) ([0bf37e6](https://github.com/netr0m/az-pim-cli/commit/0bf37e6db2e648179442326c0b101328e4fd7e82))


### Documentation

* **github:** add project guidelines ([#31](https://github.com/netr0m/az-pim-cli/issues/31)) ([5fc195b](https://github.com/netr0m/az-pim-cli/commit/5fc195bda5e78fd66b0fc996b3259d380b40f102))
* initial docs ([c315b5c](https://github.com/netr0m/az-pim-cli/commit/c315b5c44dab5102e8a7678c09e3c81d35f87a09))


### Continuous Integration

* add release-please workflow ([#36](https://github.com/netr0m/az-pim-cli/issues/36)) ([67f357d](https://github.com/netr0m/az-pim-cli/commit/67f357d1dfb1a2bc981ad257085757e59d934b90))
* add workflow triggered by merge to main ([#37](https://github.com/netr0m/az-pim-cli/issues/37)) ([4b24cf9](https://github.com/netr0m/az-pim-cli/commit/4b24cf90b8a58a5a71c36347149418b233fa038b))
