File structure

| Type        | Start | End  | Size      | Comments
| ----------- | ----- | ---- | --------- | -------------------------------- |
| FileName    | 0x00  | 0x80 | 128 bytes | IES Filename                     |
| EndFileName | 0x80  | 0x80 | 2 bytes   | End Filename Segment             |
| Meta1       | 0x84  | 0x85 | 4 bytes   |                                  | 
| Meta2       | 0x88  | 0x89 | 4 bytes   |                                  |
| Meta3       | 0x8C  | 0x8D | 4 bytes   |                                  |


Notes:

* Meta 1 - 3 seperated by 4bytes of 0x00 padding
