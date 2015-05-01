
## Objective


The purpose of this lab is to practice the thread coarsening and register tiling optimization techniques using matrix-matrix multiplication as an example.


## Instructions

Edit the file code to launch and implement a matrix-matrix multiplication kernel that uses shared memory, thread coarsening, and register tiling optimization techniques.
The first input matrix `A` is in column major format while the second matrix `B` is in row major format. 
This means that the columns and rows read from the datafile for the first input matrix are transposed.
Your code is expected to work for varying input dimensions - which may or may not be divisible by your tile size. It is a good idea to test and debug initially with examples where the matrix size is divisible by the tile size, and then try the boundary cases.



Instructions about where to place each part of the code is
demarcated by the `//@@` comment lines.
