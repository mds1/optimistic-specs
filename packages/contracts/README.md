# Optimism: Bedrock Edition - Contracts

## Install

The repo currently uses a mix of typescript tests (run with HardHat) and solidity tests (run with Forge). The project
uses the default hardhat directory structure, and all build/test steps should be run using the yarn scripts to ensure
the correct options are set.

Install node modules with yarn (v1), and Node.js (14+).

```shell
yarn
```

See installation instructions for forge [here](https://github.com/gakonst/foundry).

## Build

```shell
yarn build
```

## Running Tests


The full test suite can be executed via `yarn`:

```shell
yarn test
```

To run only typescript tests:

```shell
yarn test:hh
```

To run only solidity tests:

```shell
yarn test:forge
```
