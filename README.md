# find5
F.I.N.D. version 5

## Database setup
The `bootstrapper.sh` script will automatically create a PostGreSQL database and user for the FIND system. The database connection parameters can also be set via command line arguments:

```bash
$ ./find -h
Usage of ./find:
  -V	Print version and exit
  -dbhost string
    	database host (default "localhost")
  -dbname string
    	database name (default "finddb")
  -dbpass string
    	database password (default "dev")
  -dbport int
    	Database port (default 5432)
  -dbuser string
    	database username (default "finduser")
  -port int
    	Server port (default 5555)
```
