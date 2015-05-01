

We will be using a web based lab development and submission system through this course.
The purpose of this web system is two fold:

1. it allows you to develop CUDA applications without the need to buy and/or setup a CUDA system, and

2. it allows us to grade your submissions more efficiently, uniformly, and promptly.

This document describes the online system, walking through the pages that you will see, and describes a sample submission.
The submission system is designed to support the entire development and submission cycle for all the lab assignments.
While you have to use the submission system to receive grades for this course, we will allow you to develop locally and use the submission system after you are reasonably satisfied with your code.

## Web System

This section describes the interface and system you will be using for this course. 


### Signup

The first thing you have to do is to signup for the site.
You can use whatever username and password you want (it does not have to match the username and password used by Coursera).

![Signup1](help/imgs/signup.png "thumbnail")


### Connecting Account to Coursera

After completing the first step (or when submitting your first MP for a grade), you may be required to connect your account to Coursera (this allows us to post the grades back to Coursera).

![Signup2](help/imgs/signup2.png "thumbnail")

You will then be asked to login using your Coursera account.

![Coursera](help/imgs/coursera_login.png "thumbnail")

Once you fill in the Coursera information, you have connected your account to Coursera.

### Development Interface

Once you have signed up and logged into the website, you can start your development.
You can access the machine problems (MP) by using the **Machine Problem** tab.

Here we select the first MP from the **Machine Problems** dropdown.

![Select](help/imgs/select.png "thumbnail")

You will then be able to see the description of the MP.
This describes objective, instructions, deadlines, and the grading policy.


![SelectMPDescription](help/imgs/mp0_description.png "thumbnail")

The MP page shows the description, code, questions to be answered, previous attempts, and program history.
If you click the **Code** tab, you can start developing the MP.

![MPCode](help/imgs/mp0_code.png "thumbnail")

Once you have a solution, you click on the **Compile & Run** tab.
This will first save your work on the server and then submit it for processing (note: while code is saved automatically by the system, it is a good idea to keep a backup of the program).
In case the server is overloaded, your computation will be placed in a queue (this is a great motivation to not start the lab at the last minute).

The **Questions** tab shows questions to be answered for this MP.

![MPQuestions](help/imgs/questions.png "thumbnail")

The **Attempts** tab shows previous code submission attempts 


![MPAttempts](help/imgs/attempts.png "thumbnail")


### Insertion Points

Most of instructions for the lab are found in the code.
We use `//@@` to demarcate where code needs to be inserted.
In lab 1, for example, you see code such as:

    wbTime_start(GPU, "Copying input memory to the GPU.");
  	//@@ Copy memory to the GPU here
  	wbTime_stop(GPU, "Copying input memory to the GPU.");

This means that you need to insert a CUDA memory copy operation where the comment is located. 

### Logging and Debugging

Logging is facilitated through the use of a logging API.
The logging function `wbLog` takes a level which is either `OFF`, `FATAL`, `ERROR`, `WARN`, `INFO`, `DEBUG`, or `TRACE` and a message to be printed.
To print the value of variable `x`, for example you type `wbLog(TRACE, "The value of x = ", x)`.
Note that variables are serialized to a string when placed in the `wbLog` function, so you can place reals, integers, and strings in the logging function.

The result of the logging messages is shown in the **Attempts** tab once you submit your code.

![LoggingView](help/imgs/log.png "thumbnail")

Debugging is done in the form of logging, so you cannot place break points, etc...

### Timing Code

Logging is facilitated through the use of a timing API.
The timing functions `wbTime_start` and `wbTime_stop`.
To time a section of code, you write `wbTime_start(tag, msg)` and then `wbTime_stop(tag, msg)`, where the `tag` and `msg` must match.
In the template code given for Lab 1, for example, we do

    wbTime_start(GPU, "Allocating GPU memory.");
  	//@@ Allocate GPU memory here...
  	wbTime_stop(GPU, "Allocating GPU memory.");

the above would measure the time it takes to allocate GPU memory.
The timing result is shown in the **Attempts** tab after you submit the program.

![TimingView](help/imgs/time.png "thumbnail")

The timing information includes the tag and message you specified when you timed your code, and also which line you are timing.
Since it is important to understand the performance of your code, it is valuable to spend time to understand the output of the timer. 

### Previous Attempts

All previous attempts are saved and recorded on the server side.
To be graded for an attempt, you must click the grade button when that attempt is displayed.

# Behaviors

While developing the labs, you may encounter one or more of the following behaviors.

### No Error

![SolutionisCorrect](help/imgs/noerror.png "thumbnail")


This means that no errors were found while compiling and running the program.
Your program has been checked against the expected solution of the dataset you selected.
Note that it might be the case that the program may run correctly on one dataset and not the other.
You should check that your program runs on all datasets before submitting a program for grading.

### Solution is Incorrect

<!--
#TODO: Need screenshot!!
![SolutionisIncorrect](help/imgs/incorrect.png)
-->

This means that while the program did compile and run without errors, the solution did not match the expected results.
This problem could stem from either having an implementation error in your algorithm, or from having incorrect logic in the host code.
Logging the state of your program at different points would help you debug the problem. 

### Compilation Failed

![Compilation Failed](help/imgs/syntax.png "thumbnail")

This error message occurs when the program submitted failed to compile.
The output from the compiler is shown in the error window.
The line number of where the compilation error occurs is given in the error message.

### Attempt Limiter

<!--
![Attempt Limiter](help/imgs/limit.png "thumbnail")
-->

This error occurs when you submit more than one program within 10 seconds of each other.
To maintain stability, users can only submit one program every 10 seconds.

### Program Terminated due to Timeout

![Program Terminated](help/imgs/timeout.png "thumbnail")

The most likely cause of this error is that you have inadvertently placed an infinite loop either in your CPU or GPU code.
Part of the reason for this behavior is that the system ensures fairness (i.e. you should not hog the machine).
To ensure fairness, the system is configured to terminate long running processes.
When you see this error, you have hit that timeout limit.

### Memory Allocation Error

![Memory Allocation](help/imgs/memory.png "thumbnail")

To maintain system stability, users are not allowed to allocate too much memory.

### Sandboxing Error

![Sandbox](help/imgs/sandbox.png "thumbnail")

To maintain security, we run your program in a sandboxed mode.
This means that you are restricted to using our API functions for certain tasks, rather than using system calls or C library functions.
The sandbox error will sometimes tell you what call is being caught, although that might not always be the case.

## Machine Problem Questions

The machine problem questions are meant to allow you to think and explore your code further.
Unlike the code section, it is enough to save your answer to submit the questions for reviewing.

Answers to the questions are not graded and we encourage you to discuss them on the forums and
through your study groups.

<!--

During the peer review phase, you'll see other people's question responses and understand the different
approaches they took to implement the machine problem and what problems they encountered.

Note: if we detect that you have not answered your questions, then you'll get 0 points
for the peer review. On the second offense you'll get 0 points for the code section as well.
-->

## Development Suggestions

The following suggestions are recommended when getting started using the system

* Develop your application incrementally

* Do not modify the template code provided -- only insert code where the `//@@` demarcation is placed

* Make sure to test your results against all the datasets provided

* Do not wait until the last minute to attempt the lab.

### Further Tips

* Code is autosaved every 3 seconds and you can see the history in the history tag.
Again, do keep a backup of your programs. We suggest you backup using a private source code repository.
Free `git` private repositories are offered by Github and Bitbucket.

* You can reset the MP (get the original skeleton code) by going to
  `webgpu.com/mp/1/revert` (replacing the `1` with the mp number)
  and then back to the MP page

* Tab completion is supported in the code editor by using the `Ctrl+Space` key combination

* `Ctrl+S` saves your code along with the answers to the questions to the server

## Grading

Grading is performed based on criteria that are specific for each lab.

<!--
You will be graded not only on the code, but also on peer reviewing other people.
-->

For each attempt in the attempts tab, you'll see a *submit for grading* button.

![Attempt](help/imgs/attempt.png "thumbnail")

Clicking on that would tell the system to grade that attempt and redirect you to the grade report page.


![Grade](help/imgs/grade.png "thumbnail")

The grades are automatically sent to Coursera.
You will have to submit your grade back to Coursera by clicking the button.

<!--
## Peer Review

Two hours after the coding deadline has expired, you will have
a few days to review the work of your peers.
Clicking on the peer review button in the navigation bar will direct you to the peer review form.
There, you will be asked to review 3 of your peers.

![Peer Review](help/imgs/peer_review.png "thumbnail")

We ask you to read through the questions and code, give a score for each, and offer comments.
Once the form has been filled, you will need to submit it.

![Submit Peer Review](help/imgs/submit_peer_review.png "thumbnail")

After the peer review has been submitted, your grade page will reflect the fact that you have performed a peer
review.
You will also be able to see the reviews you peers have given you.

![Grade Peer Review](help/imgs/grade_peer_review.png "thumbnail")

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

### Peer Review Guidelines

#### Coding part

* Do you think the user handled all the boundary cases? if not which cases do you think they missed. If unsure, then you can copy their program into your coding section and run it on the datasets. Comment on what changes are needed to make the program handle all boundary cases.
* Did the user follow good programming practices?
	- Comment on the programming style: is the code indented, are the variable names used clear, were there any points in their code that were unclear and needed a comment?
	- Did the user free all memory allocated?
	- Did the user properly handle errors?
* Are there other comments you'd like to share with the user?
* Based on the above, give a grade to the program.

#### Question/Answer part

* Since some of the questions reference the program implementation, make sure to read the program before starting to review the questions.
* Read through the questions and answers commenting on whether an answer is correct or not. If not sure, ask about the question on the forums.
* Are there other comments you'd like to share with the user?
* Based on the above, give a grade to the answers.


-->


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


## Resetting Password

If you have forgotten your account password, then you can reset it by going to
the [password reset](/reset_password) page and filling in your user name and email.


![PasswordReset](help/imgs/reset_password.png "thumbnail")

You will then be emailed a link that you can use to reset your password.

![PasswordResetLink](help/imgs/password_reset_link.png "thumbnail")

Which will direct you to the password reset form

![PasswordResetLink](help/imgs/password_reset_form.png "thumbnail")


## Plagiarism

If we detect any plagiarism in the code or answers, the two parties will get 0 as their grade for the MP for the first offense.
Multiple offenses will result in a 0 for the entire course.

## System Requirements

A recent web browser is the only requirement for using and submitting labs in this course.
We have mainly tested the website using the Google Chrome browser.

## Issues/Questions/Suggestions

In case of questions, please post them to the forums.
For issues or suggestions, please use the [issue tracker](https://github.com/abduld/wb/issues).

