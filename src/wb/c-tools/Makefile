

##########################################
# INPUT
##########################################
CXX=g++
DEFINES=-DWB_USE_CUDA
DEFINES+=-DWB_USE_CUSTOM_MALLOC -DWB_USE_COURSERA
CUDA_TOOLKIT_PATH=/scr/dakkak/toolkit12.10
CXX_FLAGS=-fPIC -x c++ -O3 -I . -I $(CUDA_TOOLKIT_PATH)/include 
CXX_FLAGS+=-L $(HOME)/usr/lib -L $(HOME)/usr/cuda/lib -Wall 
CXX_FLAGS+=-I$(HOME)/usr/cuda/include -I$(HOME)/usr/include 
CXX_FLAGS+=-I/scr/dakkak/usr/cuda/include -L/scr/dakkak/usr/cuda/lib 
CXX_FLAGS+=$(DEFINES)
LIBS=#-lm -lstdc++ -lrt -lcuda -L$(HOME)/usr/lib -L/scr/dakkak/usr/cuda/lib
ARCH=$(shell uname -s)-x86_64

##########################################
##########################################

SOURCES :=  wbArg.cpp        \
			wbExit.cpp             \
			wbExport.cpp           \
			wbFile.cpp             \
			wbImage.cpp            \
			wbImport.cpp           \
			wbInit.cpp             \
			wbLogger.cpp           \
			wbMemoryManager.cpp    \
			wbPPM.cpp              \
			wbSandbox.cpp          \
			wbCUDA.cpp			   		 \
			wbSolution.cpp         \
			wbTimer.cpp


##############################################
# OUTPUT
##############################################

EXES = libwb.a libwb.so

.SUFFIXES : .o .cpp


OBJECTS = $(SOURCES:.cpp=.o)

##############################################
# OUTPUT
##############################################


.cpp.o:
	$(CXX) $(DEFINES) $(CXX_FLAGS) -c -o $@ $<


libwb.so: $(OBJECTS)
	mkdir -p $(ARCH)
	$(CXX) -fPIC -shared $(LIBS) -o $(ARCH)/$@ $(OBJECTS)

libwb.a: $(OBJECTS)
	mkdir -p $(ARCH)
	ar rcs $(ARCH)/$@ $(OBJECTS)

clean:
	rm -fr $(ARCH)
	-rm -f $(EXES) *.o *~


