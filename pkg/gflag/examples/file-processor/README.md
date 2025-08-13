# gflag Example - File Processing Tool

This example demonstrates the capabilities of the `gflag` package through a practical file processing tool that supports both POSIX-style short flags and GNU-style long flags.

## Building

```bash
go build -o gflag-example .
```

## Running the Example

### Basic Usage

```bash
# Process files with verbose output
./gflag-example -v file1.txt file2.txt
./gflag-example --verbose file1.txt file2.txt

# Process directory recursively  
./gflag-example -r /path/to/directory
./gflag-example --recursive /path/to/directory
```

### Output Formats

```bash
# JSON output to file
./gflag-example -r -f json -o results.json /path/to/dir
./gflag-example --recursive --format=json --output=results.json /path/to/dir

# CSV output
./gflag-example --format=csv /path/to/files

# Text output (default)
./gflag-example /path/to/files
```

### Combined Flags

```bash
# Combine boolean short flags
./gflag-example -vrq /path/to/dir  # verbose + recursive + quiet

# Mix short and long flags
./gflag-example -v --output=result.txt --recursive /path/to/dir

# Limit number of files processed
./gflag-example -vrq --max-count=10 --format=csv /path/to/dir
```

### Help

```bash
./gflag-example --help
./gflag-example -h
```

## Flag Formats Demonstrated

This example showcases all the flag formats supported by gflag:

### Short Flags
- `-v` (verbose)
- `-r` (recursive) 
- `-q` (quiet)
- `-h` (help)
- `-o filename` (output file)
- `-f format` (output format)
- `-c 10` (max count)

### Long Flags  
- `--verbose`
- `--recursive`
- `--quiet`
- `--help`
- `--output=filename` or `--output filename`
- `--format=json` or `--format json`
- `--max-count=10` or `--max-count 10`

### Combined Short Flags
- `-vr` (verbose + recursive)
- `-vrq` (verbose + recursive + quiet)
- `-vrf json` (verbose + recursive + format)

### Mixed Usage
- `-v --output=file.txt --recursive`
- `--verbose -r -q --format=json`

## Sample Output

### Text Format (default)
```
File Processing Results:
========================

Path: /path/to/file1.txt
Name: file1.txt
Size: 1024 bytes
Modified: 2025-08-12 10:30:45
Is Directory: false

Path: /path/to/file2.txt
Name: file2.txt  
Size: 2048 bytes
Modified: 2025-08-12 11:15:22
Is Directory: false
```

### JSON Format
```json
{
  "files": [
    {
      "path": "/path/to/file1.txt",
      "name": "file1.txt",
      "size": 1024,
      "mod_time": "2025-08-12 10:30:45", 
      "is_dir": false
    }
  ],
  "total_files": 1
}
```

### CSV Format
```csv
path,name,size,mod_time,is_dir
/path/to/file1.txt,file1.txt,1024,2025-08-12 10:30:45,false
/path/to/file2.txt,file2.txt,2048,2025-08-12 11:15:22,false
```

This example demonstrates how gflag makes it easy to create command-line tools that feel familiar to users of both POSIX and GNU tools.
