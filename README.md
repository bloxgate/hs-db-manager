# HS13 Database Manager

A tool for management of [Halo: Space Station Evolved's](https://github.com/HaloSpaceStation/HaloSpaceStation13) database.

## Features

### Compatibility

Compatible with Baystation12 *admin* and *ban* databases up through (at least) Baystation12/Baystation12@c49f12dd8657519739673d6dc3827a623fb1a809.

Compatible with HaloStation13 *whitelist* databases **after** HaloSpaceStation/HaloSpaceStation13@b811d02554966fe0d23e27fdb2253ae6d61aa1ab. Unfortunately we do not support older revisions as they expect the whitelist table to be in the old schema that is being phased out.

### Administrators

- [x] Add new administrators
- [ ] Remove administrators
- [x] Search for administrators by ckey or rank 
- [ ] Update administrators

### Bans

- [ ] Remove bans
- [x] Search for bans by:
  - [x] ckey
  - [x] Computer ID
  - [x] IP
  - [x] Admin ckey

### Alien Whitelist

- [x] Add entries
- [x] Remove entries
- [x] Search for entries

### Don't see what you need?

Open an issue and let us know!

## Running

Copy `config.toml.example` to `config.toml` and configure for your environment. `protocol` should likely be either `tcp` or `unix` (you're likely using a Unix socket to connect if you're running the database on the same machine as DreamDaemon). Then use `go run` to automatically download dependencies and run the program:

```bash
$ go run .
```

### Requirements

- Go &geq;1.13
- An internet connection to download dependencies

