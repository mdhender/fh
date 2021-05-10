# fh
Far Horizons Server

## Create a new Galaxy

The galaxy folder must contain the `setup.json` file.

All data files will be created in the galaxy folder.

```shell
$ fh create galaxy -g /path/to/galaxy/folder
```

You must run the `finish` command after creating a new galaxy.

You should run the `report` command only after `finish` completes.

### Galaxy Setup File

The `fh create galaxy` uses data from the `setup.json` file to configure the galaxy.

```json
{
  "galaxy": {
    "name": "alpha",
    "low_density": false,
    "forbid_nearby_wormholes": false,
    "minimum_distance": 10
  },
  "players": [
    {
      "email": "alderaan@example.com",
      "species_name": "Alderaan",
      "home_planet_name": "Optimus",
      "government_name": "His Majesty",
      "government_type": "Degenerated Monarchy",
      "military_level": 10,
      "gravitics_level": 1,
      "life_support_level": 1,
      "biology_level": 3
    }
  ]
}
```

The file can be deleted after the galaxy has been created.

## Finish

## Reports