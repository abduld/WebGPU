

##########################################
# INPUT
##########################################
CXX=g++
DEFINES=-DWB_USE_CUDA
DEFINES+=-DWB_USE_CUSTOM_MALLOC -DWB_USE_COURSERA
CUDA_TOOLKIT_PATH=/scr/dakkak/usr/cuda
CUDA_TOOLKIT_PATH=/usr/local/cuda
CXX_FLAGS=-fPIC -x c++ -O3 -I . -I $(CUDA_TOOLKIT_PATH)/include 
CXX_FLAGS+=-L $(HOME)/usr/lib -L $(HOME)/usr/cuda/lib -Wall  -lcuda -lcudart -lcudadevrt
CXX_FLAGS+=-I$(HOME)/usr/cuda/include -I$(HOME)/usr/include 
CXX_FLAGS+=-I$(CUDA_TOOLKIT_PATH)/include -L/usr/local/cuda/lib64
CXX_FLAGS+=-I$(CUDA_TOOLKIT_PATH)/include -L/usr/local/cuda/lib64/stubs
CXX_FLAGS+=$(DEFINES)
LIBS=-lm -lcuda -lcudart -L$(HOME)/usr/lib -L$(CUDA_TOOLKIT_PATH)/lib64 -lcudadevrt -L$(CUDA_TOOLKIT_PATH)/lib64/stubs
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
			wbSparse.cpp         \
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


libwb.so: $(SOURCES)
	mkdir -p $(ARCH)
	$(CXX) -fPIC -shared $(LIBS) -o $(ARCH)/$@ $(SOURCES) $(DEFINES) $(CXX_FLAGS)

libwb.a: $(OBJECTS)
	mkdir -p $(ARCH)
	ar rcs $(ARCH)/$@ $(OBJECTS)

clean:
	rm -fr $(ARCH)
	-rm -f $(EXES) *.o *~


