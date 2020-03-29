/*************************************************************************
  > File Name: gfs.h
  > Author:perrynzhou 
  > Mail:perrynzhou@gmail.com 
  > Created Time: Fri 06 Mar 2020 04:50:39 PM CST
 ************************************************************************/

#ifndef _GFS_H
#define _GFS_H
#include <glusterfs/api/glfs.h>
#include <glusterfs/api/glfs-handles.h>
#include <stdint.h>
typedef struct gfs_api_t
{
  glfs_t *fs;
  char *suffix;
  uint64_t index;
  char *index_path;
  char *index_name;
  char *buffer;
  uint64_t buffer_size;
  uint64_t max;
  uint64_t cur;
} gfs_api;
glfs_t *glfs_create(const char *volume, const char *addr, int port);
glfs_t *glfs_destroy(glfs_t *gt);
gfs_api *gfs_api_create(glfs_t *fs, const char *suffix, const char *index_path, const char *index_name, uint64_t buffer_size,uint64_t max);
int gfs_api_search(gfs_api *api,  const char *root);
int gfs_api_read(gfs_api *api, const char *path);
void gfs_api_destroy(gfs_api *api);

#endif
