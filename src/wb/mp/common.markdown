
## Suggestions

* The system's autosave feature is not an excuse to not backup your code and answers to your questions regularly.

* If you have not done so already, read the [tutorial](/help)

* Do not modify the template code provided -- only insert code where the `//@@` demarcation is placed

* Develop your application incrementally

* Do not wait until the last minute to attempt the lab. 

* If you get stuck with boundary conditions, grab a pen and paper. It is much easier to figure out the boundary conditions there.

* Implement the serial CPU version first, this will give you an understanding of the loops

* Get the first dataset working first. The datasets are ordered so the first one is the easiest to handle

* Make sure that your algorithm handles non-regular dimensional inputs (not square or multiples of 2).
The slides may present the algorithm with nice inputs, since it minimizes the conditions.
The datasets reflect different sizes of input that you are expected to handle

* Make sure that you test your program using all the datasets provided (the datasets can be selected using the dropdown next to the submission button)

* Check for errors: here is an example function of a macro `wbCheck` that you can use to check for errors while programming CUDA:

        #define wbCheck(stmt) do {                                                    \
                cudaError_t err = stmt;                                               \
                if (err != cudaSuccess) {                                             \
                    wbLog(ERROR, "Failed to run stmt ", #stmt);                       \
                    wbLog(ERROR, "Got CUDA error ...  ", cudaGetErrorString(err));    \
                    return -1;                                                        \
                }                                                                     \
            } while(0)


An example usage is `wbCheck(cudaMalloc(...))`

## Plagiarism

Plagiarism will not be tolerated.
The first offense will result in the two parties getting a 0 for the machine problem.
Second offense results in a 0 for the course.

## Grading

Grading is performed based on criteria that are specific for each lab.
You will be graded not only on the code, but also on peer reviewing other people.
For each attempt in the attempts tab, you'll see a *submit for grading* button.

![Attempt](/help/imgs/attempt.png "thumbnail")

Clicking on that would tell the system to grade that attempt and redirect you to the grade report page.


![Grade](/help/imgs/grade.png "thumbnail")

The grades are not automatically sent to Coursera.
You will have to submit your grade back to Coursera by clicking the button.

## Peer Review

Two hours after the coding deadline has expired, you will have
a few days to review the work of your peers.
Clicking on the peer review button in the navigation bar will direct you to the peer review form.
There, you will be asked to review 3 of your peers.

![Peer Review](/help/imgs/peer_review.png "thumbnail")

We ask you to read through the questions and code, give a score for each, and offer comments.
Once the form has been filled, you will need to submit it.

![Submit Peer Review](/help/imgs/submit_peer_review.png "thumbnail")

After the peer review has been submitted, your grade page will reflect the fact that you have performed a peer
review.
You will also be able to see the reviews you peers have given you.

![Grade Peer Review](/help/imgs/grade_peer_review.png "thumbnail")

At this point you will want to post the peer review grade back to Coursera.


The reason we incorporated peer review in the course is three fold.
First, if you were not able to get the correct code for the MP, it would allow you the opportunity to examine other people's code and get feedback as to why your code did not work (with the idea that this would help you in subsequent MPs).
Second, it makes you familiar with approaches to solve the problem.
We believe that reading *good* or *bad* code is just as important as writing them.
And last, since you are partly reviewed on code neatness, it forces you to add adequate comments, use sensible variable names, etc...

Since peer review is an essential part of the course, we will be weighing it heavily.
We expect you to give constructive feedback that would help your peer in progressing through the course.
If we detect that you have not been giving adequate reviews,  you will get a 0 for the peer review section for the first offense.
The second offense will result in a 0 for the machine problem.


## Local Development

While not required, the library used throughout the course can be downloaded from [Github](https://github.com/abduld/libwb).
The library does not depend on any external library (and should be cross platform), you can use `make` to generate the shared object file (further instructions are found on the Github page).
Linking against the library would allow you to get similar behavior to the web interface (minus the imposed limitations).
Once linked against the library, you can launch the your program as follows:

    ./program -e <expected_output_file> -i <input_file_1>,<input_file_2> -o <output_file> -t <type>

The `<expected_output_file>` and `<input_file_n>` are the input and output files provided in the dataset.
The `<output_file>` is the location you'd like to place the output from your program.
The `<type>` is the output file type: `vector`, `matrix`, or `image`.
If an MP does not expect an input or output, then pass `none` as the parameter.

In case of issues or suggestions with the library, please report them
through the [issue tracker](https://github.com/abduld/libwb/issues) on Github.


## Issues/Questions/Suggestions

In case of questions, please post them to the forums.
For issues or suggestions, please use the [issue tracker](https://github.com/abduld/wb/issues).

