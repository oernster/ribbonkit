// The C half of gtkmain. It lives apart from the Go file because a Go file that exports a
// function to C may only declare C functions, never define them.

#include <gtk/gtk.h>
#include "_cgo_export.h"

// gtkmain_run is the loop's side of Do: it runs the Go function the handle holds, once.
static gboolean gtkmain_run(gpointer data)
{
    gtkmainRun((GoUintptr)data);
    return G_SOURCE_REMOVE;
}

// gtkmain_invoke runs the handle's function on the default main context: at once when this thread
// owns it, otherwise queued for the loop.
void gtkmain_invoke(guintptr handle)
{
    g_main_context_invoke(NULL, gtkmain_run, (gpointer)handle);
}

// gtkmain_init opens GTK without arguments, answering FALSE where no display can be opened.
gboolean gtkmain_init(void)
{
    return gtk_init_check(NULL, NULL);
}
