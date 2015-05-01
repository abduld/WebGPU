
## Objective

Implement a basic dense matrix multiplication routine that multiplies an m\*k matrix by a k\*n matrix to produce an m\*n matrix. See lecture slides for details on the mathematical function. It is a good idea to test and debug initially with small input dimensions and dimensions divisible by the block size then make your test cases more complicated as you go. Your code is expected to work for varying input dimensions – which may or may not be divisible by your block size and may or may not be square – so don’t forget to pay attention to boundary conditions. We provide you with a number of test cases, in increasing order of difficulty.

## Prerequisites

Before starting this lab, make sure that:

* You have completed the "Vector Addition" Lab


## Instruction

Edit the code in the code tab to perform the following:

- allocate device memory
- copy host memory to device
- initialize thread block and kernel grid dimensions
- invoke CUDA kernel
- copy results from device to host
- deallocate device memory
- implement CUDA kernel

Instructions about where to place each part of the code is
demarcated by the `//@@` comment lines.

## Grading
Your submission will be graded based on the following criteria.

• Functionality/knowledge: 75% 

o Works for square matrices 

o Works for multiples of 16 \* 16 tiles 

o Works for non-square matrices

o Works for arbitrary sizes (non-multiple of 16) matrices 

o Correct usage of CUDA library calls and C extensions 

• Answers to questions: 25% 

o Correct answers to questions 

o Sufficient and clear explanation is given (a sentence or two) 