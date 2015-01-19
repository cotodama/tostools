File structure

| Type        | Start | End  | Size      | Comments
| ----------- | ----- | ---- | --------- | -------------------------------- |
| FileName    | 0x00  | 0x80 | 128 bytes | IES Filename                     |
| EndFileName | 0x80  | 0x80 | 2 bytes   | End Filename Segment             |
| Meta1       | 0x84  | 0x85 | 4 bytes   | Reverse Endian                   | 
| Meta2       | 0x88  | 0x89 | 4 bytes   | Reverse Endian                   |
| FileSize    | 0x8C  | 0x8D | 4 bytes   | Reverse Endian                   |
| EndMeta     | 0x90  | 0x90 | 2 bytes   | End File Meta Segment            |


Notes:

* Meta 1 - 3 seperated by 4bytes of 0x00 padding
* Meta 1 & 2 and FileSize are reverse endian
