/*************************************************************************
  > File Name: list.c
  > Author:perrynzhou 
  > Mail:perrynzhou@gmail.com 
  > Created Time: Sun 24 Nov 2019 10:53:08 AM CST
 ************************************************************************/

#include "list.h"
#include <stdio.h>
#include <stdlib.h>
list_node *list_node_create(void *value)
{
  list_node *node = NULL;
  if (value != NULL)
  {
    node = (list_node *)calloc(1, sizeof(list_node));
    if (node != NULL)
    {
      node->prev = node->next = NULL;
      node->value = value;
    }
  }
  return node;
}
inline void list_node_destroy(list_node *node)
{
  if (node != NULL)
  {
    node->prev = node->next = NULL;
    free(node);
    node = NULL;
  }
}
list *list_create()
{
  list *lt = (list *)calloc(1, sizeof(list));
  if (lt != NULL)
  {
    lt->head = lt->tail = NULL;
    lt->size = 0;
  }
  return lt;
}
int list_add(list *lt, list_node *node)
{
  if (lt != NULL)
  {
    if (lt->size == 0)
    {
      lt->head = lt->tail = node;
    }
    else
    {
      lt->tail->next = node;
      node->prev = lt->tail;
      lt->tail = node;
    }
    __sync_fetch_and_add(&lt->size, 1);
    return 0;
  }
  return -1;
}
int list_del(list *lt, list_node *node)
{

  if (lt != NULL && lt->size > 0)
  {
    if (node->prev != NULL)
    {
      node->prev->next = node->next;
    }
    else
    {
      lt->head = node->next;
    }
    if (node->next != NULL)
    {
      node->next->prev = node->prev;
    }
    else
    {
      lt->tail = node->prev;
    }
    list_node_destroy(node);
    __sync_fetch_and_sub(&lt->size, 1);
  }
}
void list_destroy(list *lt)
{
  if (lt != NULL && lt->size > 0)
  {
    list_node *head = lt->head;
    list_node *tail = lt->tail;
    lt->size = 0;
    if (head != NULL)
    {
      while (head != tail)
      {
        list_node *next = head->next;
        list_node *prev = tail->prev;
        list_node_destroy(head);
        list_node_destroy(tail);
        head = next;
        tail = prev;
        if (head == NULL || tail == NULL)
        {
          break;
        }
      }
    }
    free(lt);
    lt = NULL;
  }
}