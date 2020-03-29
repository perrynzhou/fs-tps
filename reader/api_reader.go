package reader

/*
#include "../api/list.h"
#include "../api/gfs.h"
*/
import "C"
import (
	"fmt"
	"glusterfs-tps/conf"
	"unsafe"
)
type ApiReader struct {
	api *C.gfs_api
}
func NewApiReader(cf *conf.Conf) (*ApiReader,error) {
	addr :=C.CString(cf.Address)
	port :=C.int(cf.Port)
	vol := C.CString(cf.Volume)
	defer      C.free(unsafe.Pointer(addr))
	defer      C.free(unsafe.Pointer(vol))
	gf :=C.glfs_create(vol,addr,port);
	if gf == nil  {
		return nil,fmt.Errorf("init glfs_create failed")
	}
	indexPath :=C.CString(cf.IndexPath)
	suffix :=C.CString(cf.Suffix)
	indexName :=C.CString(cf.IndexName)
	defer      C.free(unsafe.Pointer(indexPath))
	defer      C.free(unsafe.Pointer(indexName))
	bufferSize :=C.ulonglong(cf.BufferSize)
	max := cf.Count
	if max < defaultFilecount {
		max = defaultFilecount
	}
	api :=C.gfs_api_create(gf,suffix,indexPath,indexName,bufferSize,C.ulonglong(max))
	if api == nil {
		C.glfs_destroy(gf);
		return nil,fmt.Errorf("gfs_api_create init failed")
	}
	return &ApiReader{
		api:api,
	},nil
}
func (apiReader *ApiReader)Walk(path string) int {
	root := C.CString(path)
	defer C.free(root)
    return int(C.gfs_api_search(apiReader.api,root))
}
func (apiReader *ApiReader) Read(path string, flag bool, handle func(b []byte) error) error {
	cPath := C.CString(path)
	ret := C.gfs_api_read(apiReader.api,cPath)
	if int(ret)!=0 {
		return fmt.Errorf("gfs_api_read ret:%d",int(ret))
	}
	return nil
}

