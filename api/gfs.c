/*************************************************************************
  > File Name: gfs.c
  > Author:perrynzhou 
  > Mail:perrynzhou@gmail.com 
  > Created Time: Fri 06 Mar 2020 04:50:44 PM CST
 ************************************************************************/

#include "gfs.h"
#include "list.h"
#include <sys/types.h>
#include <sys/stat.h>
#include <fcntl.h>
#include <glusterfs/api/glfs.h>
#include <glusterfs/api/glfs-handles.h>
glfs_t *glfs_create(const char *volume, const char *addr, int port)
{
  glfs_t *fs = NULL;
  if (volume != NULL && addr != NULL)
  {
    fs = glfs_new(volume);
    if (fs != NULL)
    {
      if (glfs_set_volfile_server(fs, "tcp", addr, port) != 0 && glfs_init(fs) != 0)
      {
        glfs_fini(fs);
        fs = NULL;
      }
    }
  }
  return fs;
}
glfs_t *glfs_destroy(glfs_t *gt){
  if(gt!=NULL) {
    glfs_fini(gt);
    gt=NULL;
  }
}
gfs_api *gfs_api_create(glfs_t *fs, const char *suffix, const char *index_path, const char *index_name, uint64_t buffer_size, uint64_t max)
{
  if (fs == NULL || index_name == NULL || index_path == NULL)
  {
    return NULL;
  }
  gfs_api *api = malloc(sizeof(gfs_api));
  assert(api != NULL);
  api->fs = fs;
  api->buffer = (char *)malloc(buffer_size);
  assert(api->buffer != NULL);
  api->buffer_size = buffer_size;
  api->index_name = strdup(index_name);
  api->suffix = strdup(suffix);
  api->index_path = strdup(index_path);
  api->index = 0;
  api->max = max;
  api->cur = 0;
  return api;
}
static int gfs_api_create_fd(gfs_api *api)
{
  char buffer[256] = {'\0'};
  snprintf((char *)&buffer, "%s/%s.%d", api->index_path, api->index_name, api->index);
  int fd = open((char *)&buffer, O_CREAT | O_APPEND | O_TRUNC | O_RDWR);
  assert(fd != -1);
  __sync_fetch_and_add(&api->index, 1);
  return fd;
}
int gfs_api_read(gfs_api *api, const char *path)
{
  int ret = -1;
  glfs_fd_t *fd = glfs_h_open(api->fs, path, O_WRONLY);
  if (fd != NULL)
  {

    ret = glfs_read(fd, api->buffer, api->buffer_size, 0);
  }
out:
  if (fd != NULL)
  {
    glfs_close(fd);
  }
  return ret;
}
int gfs_api_search(gfs_api *api, const char *path)
{
  {
    size_t len = 0;
    if (api->suffix != NULL)
    {
      len = strlen(api->suffix);
    }
    list *lt = list_create();
    char *root = strdup(path);
    assert(root != NULL);
    list_node *first = list_node_create(root);
    list_add(lt, first);
    list_node *cur = lt->head;
    first = lt->head;
    int ret = -1;
    int wfd = gfs_api_create_fd(api);
    while (cur != NULL)
    {
      char buf[512] = {'\0'};
      struct dirent *dt = NULL;
      glfs_fd_t *fd = glfs_opendir(api->fs, (char *)cur->value);
      while (glfs_readdir_r(fd, (struct dirent *)buf, &dt), dt)
      {
        if (dt->d_type == DT_REG || dt->d_type == DT_LNK)
        {

          size_t d_name_len = strlen(dt->d_name);
          if (len > 0)
          {
            fprintf(stdout, "suffix:%s-------%s\n", api->suffix, dt->d_name + d_name_len - len);
          }
          if (len == 0 || (len > 0 && (len < d_name_len) && strncmp(dt->d_name + d_name_len - len, api->suffix, len)) == 0)
          {
            if (api->cur >= api->max)
            {
              close(wfd);
              wfd = gfs_api_create_fd(api);
            }
            char buffer[4096] = {'\0'};
            snprintf((char *)&buffer, 255, "%s/%s", cur->value, dt->d_name);
            int w_sz = write(wfd, (char *)&buffer, strlen((char *)&buffer));
            assert(w_sz > 0);
          }
        }
        else if (dt->d_type == DT_DIR)
        {
          if (strncmp(dt->d_name, ".", len) == 0 || strncmp(dt->d_name, "..", len) == 0)
          {
            continue;
          }
          char buffer[255] = {'\0'};
          snprintf(&buffer, 255, "%s/%s", cur->value, dt->d_name);
          list_node *node = list_node_create(strdup((char *)&buffer));
          assert(node != NULL);
          list_add(lt, node);
        }
      }
      glfs_closedir(fd);
      cur = cur->next;
    }
    if (wfd != -1)
    {
      close(wfd);
    }
    cur = lt->tail;
    while (cur != NULL)
    {
      list_node *prev = cur->prev;
      free(cur->value);
      list_node_destroy(cur);
      cur = prev;
    }
    lt->head = lt->tail = NULL;
    if (lt != NULL)
    {
      list_destroy(lt);
    }
  }
  return 0;
}
void gfs_api_destroy(gfs_api *api)
{
  if (api != NULL)
  {
    free(api->buffer);
    free(api->index_name);
    free(api->index_path);
    if (api->suffix != NULL)
    {
      free(api->suffix);
      api->suffix = NULL;
    }
    glfs_destroy(api->fs);
    free(api);
    api = NULL;
  }
}
#ifdef  TEST
int main(int argc,char *argv[]){
    char *add = "127.0.0.1";
    int port = 24007;
    char *vol = "dht_vol";
    
}
#endif