Header structure

| Type        | Start | End  | Size      | Comments
| ----------- | ----- | ---- | --------- | -------------------------------- |
| FileName    | 0x00  | 0x80 | 128 bytes | IES Filename                     |
| DataPtr     | 0x84  | 0x85 | 4 bytes   | Little Endian                    | 
| ResPtr      | 0x88  | 0x89 | 4 bytes   | Little Endian                    |
| FileSize    | 0x8C  | 0x8D | 4 bytes   | Little Endian                    |


Notes:

* the strings characters are scrambled with an xor key (0x1)  
* the table structure is defined first  
* the table contents are defined at the end (packed in the format of structure)  

* resources start at eof - res  
* data starts at resources - dataOffset  

bgm.ies:  

    - fileName: Bgm 
    - dataOffset: 816 
    - resOffset: 10769 
    - eofOffset: 11741 
    - resPtr: 972 
    - dataPtr: 156 
    - numRows: 128
    - numFormats: 6


    int str int[] str[]

[{ClassID ClassID 0} {ClassName ClassName 1} {Composer Composer 1} {Copyrights Copyrights 1} {FileName FileName 1} {Staff Staff 1}]