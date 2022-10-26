# EBing_Wallpaper

This small tool, writen in go, will change the wallpaper to the Bing Wallpaper of the day.

This is my first project written in go.

## Usage

EBing_Wallpaper [-hkqsvV] [-p value]

| Short | Long      | Description                                               |
| ----- | --------- | --------------------------------------------------------- |
| -d    | --daemon  | Start in the background.                                  |
| -h    | --help    | Display this help                                         |
| -k    | --keep    | Keep the tmp files, do not delete them.                   |
| -p    | --past    | is the previous [n] wallpaper from Bing. Max 7 days back. |
| -q    | --quiet   | No output, all output quiet.                              |
| -s    | --skip    | Skip the Enlightenment desktop check.                     |
| -v    | --verbose | Verbose output                                            |
| -V    | --version | Display version                                           |

## Daemon mode

In daemon mode, it will get and set the background every day at 09:00 CEST.

This time is not (yet) variable.
