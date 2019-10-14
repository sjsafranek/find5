
TODO
python_ai starts up with default datafolder
switch to json config files
allow api client to set timestamps


5.0.5
created separate directory to track database schema and patches
added logging directory command line argument
cleanup of python ai
removed ai datafolder requirement in api calls

5.0.4
fixed tcp pool for retry on golang side
no longer writes csv files to disk
sends gzipped csv with base64 encoding through json message over tcp connection

5.0.3
golang api with method callbacks

5.0.2
database golang api
fixed database constraints for measurements table

5.0.1
database structure
