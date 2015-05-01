

#include "stdio.h"
#include "stdlib.h"
#include "assert.h"
#include "string.h"
#include "sys/stat.h"

static void compute(float *output, float *inputValues, float *inputPositions,
                    int numInputs, int gridSize) {
  // Compute output
  for (int inIdx = 0; inIdx < numInputs; ++inIdx) {
    for (int outIdx = 0; outIdx < gridSize; ++outIdx) {
      float dist = inputPositions[inIdx] - ( float )outIdx;
      if (dist == 0) {
        continue;
      }
      output[outIdx] +=
          (inputValues[inIdx] * inputValues[inIdx]) / (dist * dist);
    }
  }
  return;
}

static float *generate_data(int len, int max) {
  float *res = ( float * )malloc(len * sizeof(float));
  for (int i = 0; i < len; i++) {
    res[i] = ((( float )rand()) / RAND_MAX) * max;
  }
  return res;
}

static char *strjoin(const char *s1, const char *s2) {
  char *result = ( char * )malloc(strlen(s1) + strlen(s2) + 1);
  strcpy(result, s1);
  strcat(result, s2);
  return result;
}

static char base_dir[] = "./data";

static void write_data(char *file_name, float *data, int len) {
  FILE *handle = fopen(file_name, "wb");
  fprintf(handle, "%d\n", len);
  for (int ii = 0; ii < len; ii++) {
    fprintf(handle, "%.2f", *data++);
    if (ii != len - 1) {
      fprintf(handle, "\n");
    }
  }
  fclose(handle);
}

static void write_data(char *file_name, int flag) {
  FILE *handle = fopen(file_name, "w");
  fprintf(handle, "%d\n", flag);
  fflush(handle);
  fclose(handle);
}

static void create_dataset(int mode, int num, int len, int max, int gridSize) {
  char dir_name[1024];
  num = mode * 10 + num;

  sprintf(dir_name, "%s/%d", base_dir, num);

  mkdir(dir_name, 0777);

  char *mode_file_name = strjoin(dir_name, "/mode.raw");
  char *input0_file_name = strjoin(dir_name, "/input0.raw");
  char *input1_file_name = strjoin(dir_name, "/input1.raw");
  char *grid_size_file_name = strjoin(dir_name, "/grid_size.raw");
  char *output_file_name = strjoin(dir_name, "/output.raw");

  float *input0_data = generate_data(len, max);
  float *input1_data = generate_data(len, gridSize);
  float *output_data = ( float * )calloc(sizeof(float), gridSize);

  compute(output_data, input0_data, input1_data, len, gridSize);

  write_data(mode_file_name, mode);
  write_data(input0_file_name, input0_data, len);
  write_data(input1_file_name, input1_data, len);
  write_data(grid_size_file_name, gridSize);
  write_data(output_file_name, output_data, gridSize);
}

enum Mode {CPU_NORMAL = 1, GPU_NORMAL, GPU_CUTOFF,
        GPU_BINNED_CPU_PREPROCESSING, GPU_BINNED_GPU_PREPROCESSING};

static void create_dataset(int num, int len, int max, int gridSize) {
    create_dataset(CPU_NORMAL, num, len, max, gridSize);
    create_dataset(GPU_NORMAL, num, len, max, gridSize);
    create_dataset(GPU_CUTOFF, num, len, max, gridSize);
    create_dataset(GPU_BINNED_CPU_PREPROCESSING, num, len, max, gridSize);
    create_dataset(GPU_BINNED_GPU_PREPROCESSING, num, len, max, gridSize);
}

int main() {
  create_dataset(0, 60, 1, 60);
  create_dataset(1, 600, 1, 100);
  create_dataset(2, 3 * 201, 1, 201);
  create_dataset(3, 4 * 100 + 9, 1, 160);
  create_dataset(4, 2 * 210 - 1, 1, 100);
  create_dataset(5, 4 * 2011 + 21, 1, 201);
  create_dataset(6, 1440, 1, 443);
  create_dataset(7, 4 * 100, 1, 200);
  create_dataset(8, 3 * 232, 1, 232);
  return 0;
}
