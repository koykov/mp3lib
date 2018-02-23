// This package is a lightweight version of https://github.com/hwch/go-mp3-player
package mp3lib

/*
#cgo pkg-config: gstreamer-0.10
#include <gst/gst.h>
gboolean bus_call(GstBus *bus, GstMessage *msg, gpointer data)
{
	GMainLoop *loop = (GMainLoop *)data;
	gchar *debug;
	GError *error;
	switch (GST_MESSAGE_TYPE(msg)) {
	case GST_MESSAGE_EOS:
		g_main_loop_quit(loop);
		break;
	case GST_MESSAGE_ERROR:
		gst_message_parse_error(msg,&error,&debug);
		g_free(debug);
		//g_printerr("Bus error: %s\n",error->message);
		g_error_free(error);
		g_main_loop_quit(loop);
		break;
	default:
		break;
	}
	return TRUE;
}
static GstBus *pipeline_get_bus(void *pipeline)
{
	return gst_pipeline_get_bus(GST_PIPELINE(pipeline));
}
static void bus_add_watch(void *bus, void *loop)
{
	gst_bus_add_watch(bus, bus_call, loop);
	gst_object_unref(bus);
}
static void set_path(void *play, gchar *path)
{
	g_object_set(G_OBJECT(play), "uri", path, NULL);
}
static void object_unref(void *pipeline)
{
	gst_object_unref(GST_OBJECT(pipeline));
}
static void media_ready(void *pipeline)
{
	gst_element_set_state(pipeline, GST_STATE_READY);
}
static void media_play(void *pipeline)
{
	gst_element_set_state(pipeline, GST_STATE_PLAYING);
}
static void media_pause(void *pipeline)
{
	gst_element_set_state(pipeline, GST_STATE_PAUSED);
}
static void media_stop(void *pipeline)
{
	gst_element_set_state(pipeline, GST_STATE_NULL);
}
static void set_mute(void *play)
{
	g_object_set(G_OBJECT(play), "mute", FALSE, NULL);
}
static void set_volume(void *play, int vol)
{
	int ret = vol % 101;
	g_object_set(G_OBJECT(play), "volume", ret/10.0, NULL);
}
static void media_seek(void *pipeline, gint64 pos)
{
	gint64 cpos;
	GstFormat gformat = GST_FORMAT_TIME;
	gst_element_query_position (pipeline, &gformat, &cpos);
	cpos += pos*1000*1000*1000;
	if (!gst_element_seek (pipeline, 1.0, GST_FORMAT_TIME, GST_SEEK_FLAG_FLUSH,
                         GST_SEEK_TYPE_SET, cpos,
                         GST_SEEK_TYPE_NONE, GST_CLOCK_TIME_NONE)) {
    		g_print ("Seek failed!\n");
    	}
}
*/
import "C"

import (
	"container/list"
	"fmt"
	"unsafe"
)

var g_list *list.List
var g_isQuit bool = false
var pipeline *C.GstElement
var bus *C.GstBus
var g_volume_size int = 10

func GString(s string) *C.gchar {
	return (*C.gchar)(C.CString(s))
}

func GFree(s unsafe.Pointer) {
	C.g_free(C.gpointer(s))
}

func PlayProcess(url string) {
	var loop *C.GMainLoop

	g_list = list.New()
	C.gst_init((*C.int)(unsafe.Pointer(nil)),
		(***C.char)(unsafe.Pointer(nil)))
	loop = C.g_main_loop_new((*C.GMainContext)(unsafe.Pointer(nil)),
		C.gboolean(0))
	g_list.PushBack(url)

	sig_out := make(chan bool)
	e := g_list.Front()

	v0 := GString("playbin")
	v1 := GString("play")
	pipeline = C.gst_element_factory_make(v0, v1)
	GFree(unsafe.Pointer(v0))
	GFree(unsafe.Pointer(v1))
	bus = C.pipeline_get_bus(unsafe.Pointer(pipeline))
	if bus == (*C.GstBus)(nil) {
		fmt.Println("GstBus element could not be created. Exiting.")
		return
	}
	C.bus_add_watch(unsafe.Pointer(bus), unsafe.Pointer(loop))

	go func(sig_quit chan bool) {
		i := 0
		for !g_isQuit {
			if i != 0 {
				C.media_ready(unsafe.Pointer(pipeline))
				C.media_play(unsafe.Pointer(pipeline))
			}
			C.g_main_loop_run(loop)
			C.media_stop(unsafe.Pointer(pipeline))

			fpath, ok := e.Value.(string)
			if ok {
				v2 := GString(fpath)
				C.set_path(unsafe.Pointer(pipeline), v2)
				GFree(unsafe.Pointer(v2))

			} else {
				break
			}
			i++
		}

		C.object_unref(unsafe.Pointer(pipeline))

	}(sig_out)

	fpath, ok := e.Value.(string)
	if ok {
		v2 := GString(fpath)
		C.set_path(unsafe.Pointer(pipeline), v2)
		GFree(unsafe.Pointer(v2))

		C.media_ready(unsafe.Pointer(pipeline))
		C.media_play(unsafe.Pointer(pipeline))
	}
	C.g_main_loop_quit(loop)
}

func PauseProcess() {
	C.media_pause(unsafe.Pointer(pipeline))
}

func ResumeProcess() {
	C.media_play(unsafe.Pointer(pipeline))
}

func StopProcess() {
	C.media_stop(unsafe.Pointer(pipeline))
}

func MuteProcess() {
	g_volume_size = 0
	C.set_volume(unsafe.Pointer(pipeline), C.int(g_volume_size))
}

func UnmuteProcess() {
	g_volume_size = 10
	C.set_volume(unsafe.Pointer(pipeline), C.int(g_volume_size))
}