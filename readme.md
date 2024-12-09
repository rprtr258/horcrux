# horcrux <img src="https://static.wikia.nocookie.net/minecraft_gamepedia/images/3/33/Chorus_Fruit_JE2_BE2.png/revision/latest/thumbnail/width/360/height/360?cb=20190505051041" width="30" height="30" />

## Goal
Main purpose of this utility is to split input file into `N` chunks, of which any `K` can be used to restore original file. Any `K-1` chunks or less will not be enough.

That means `N >= 2` AND `1 < K <= N`. `N=1` or `K=1` does not make sense, just copy your file for that. `K=N` means splitting input file into even chunks.

## Usage
```bash
# split into 4 horcruxes, any 3 can be used to restore
go run . split -n 4 -k 3 diary.txt

# diary.txt.{0,1,2,3}.horcrux files are created

# join 0, 1 and 3 horcruxes
go run . join -o out.txt diary.txt.{0,1,3}.horcrux
```

## Lore
Basic idea from [jesseduffield/horcrux](https://github.com/jesseduffield/horcrux) which uses key-splitting and encryption to achieve similar goal.

This project just splits source file, which makes every horcrux of size `1-(K-1)/N` of original file size (exercise to the reader: prove that).

For example, splitting example [diary.txt](./diary.txt) file with `N=4` and `K=3` will result in following files:
```bash
$ ls -la *txt*
-rw-r--r-- 1 rprtr258 rprtr258 1795 Dec  9 01:34 diary.txt
-rw-r--r-- 1 rprtr258 rprtr258 1021 Dec  9 03:46 diary.txt.0.horcrux
-rw-r--r-- 1 rprtr258 rprtr258 1021 Dec  9 03:46 diary.txt.1.horcrux
-rw-r--r-- 1 rprtr258 rprtr258 1022 Dec  9 03:46 diary.txt.2.horcrux
-rw-r--r-- 1 rprtr258 rprtr258 1022 Dec  9 03:46 diary.txt.3.horcrux
-rw-r--r-- 1 rprtr258 rprtr258 1795 Dec  9 03:46 output.txt
```

Notice each horcrux is approximately `1-(K-1)/N = 1-(3-1)/4 = 1/2` of original file size.
