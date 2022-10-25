# EBing_Wallpaper

This small tool, writen in go, will change the wallpaper to the Bing Wallpaper of the day.

This is my first project written in go.

## Usage

EBing_Wallpaper [-hkqsvV] [-p value]

| Short | Long      | Description                                               |
| ----- | --------- | --------------------------------------------------------- |
| -h    | --help    | Display this help                                         |
| -k    | --keep    | Keep the tmp files, do not delete them.                   |
| -p    | --past    | is the previous [n] wallpaper from Bing. Max 7 days back. |
| -q    | --quiet   | No output, all output quiet.                              |
| -s    | --skip    | Skip the Enlightenment desktop check.                     |
| -v    | --verbose | Verbose output                                            |
| -V    | --version | Display version                                           |

## Crontab

You can't use EBing_Wallpaper in a cronjob yet.
I will try to find a solution for this problem.
