/*************************************************************************
  > File Name: list.h
  > Author:perrynzhou 
  > Mail:perrynzhou@gmail.com 
  > Created Time: Sun 24 Nov 2019 10:52:54 AM CST
 ************************************************************************/

#ifndef _LIST_H
#define _LIST_H
#include <stdio.h>
typedef struct list_node_t
{
  struct list_node_t *next;
  struct list_node_t *prev;
  void *value;
} list_node;
typedef struct list_t
{
  list_node *head;
  list_node *tail;
  size_t size;
} list;
//list_node op
list_node *list_node_create(void *value);
void list_node_destroy(list_node *node);
//list op
list *list_create();
int list_add(list *lt,list_node *node);
int list_del(list *lt,list_node *node);
void list_destroy(list *lt);
#endif
