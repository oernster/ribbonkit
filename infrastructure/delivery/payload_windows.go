package delivery

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/oernster/ribbonkit/infrastructure/setup"
)

// packingSuffix follows the archive's name while it is packed.
const packingSuffix = ".packing"

// errPayloadFlags is answered when Payload is not told where the application, the licence or the
// archive is.
var errPayloadFlags = errors.New("-app, -licence and -out are all needed")

// Payload packs the setup program's payload for product: every file of the built application, then
// the licence beside it. The archive is written beside its target and moved into place only once
// whole, so a packing that fails leaves what was there before. Its flags:
//
//	-app build/bin -licence LICENSE -out installer/payload.zip
func Payload(args []string, out io.Writer, product setup.Product) error {
	flags := flag.NewFlagSet("payload", flag.ContinueOnError)
	flags.SetOutput(out)
	app := flags.String("app", "", "the folder holding the built application")
	licence := flags.String("licence", "", "the licence to carry beside it")
	archive := flags.String("out", "", "the archive to write")
	if err := flags.Parse(args); err != nil {
		return err
	}
	if *app == "" || *licence == "" || *archive == "" {
		return errPayloadFlags
	}
	if err := writeWhole(*archive, setup.Payload{App: *app, Exe: product.Exe(), Licence: *licence}); err != nil {
		return err
	}
	fmt.Fprintf(out, "%s holds the application and its licence\n", *archive)
	return nil
}

// writeWhole packs into a file beside archive, then moves it into place.
func writeWhole(archive string, payload setup.Payload) error {
	packing := archive + packingSuffix
	file, err := os.OpenFile(packing, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, fileMode)
	if err != nil {
		return fmt.Errorf("creating %s: %w", packing, err)
	}
	err = setup.Pack(file, payload)
	if closeErr := file.Close(); err == nil && closeErr != nil {
		err = fmt.Errorf("writing %s: %w", packing, closeErr)
	}
	if err != nil {
		_ = os.Remove(packing)
		return err
	}
	if err := os.Rename(packing, archive); err != nil {
		return fmt.Errorf("moving %s into place: %w", archive, err)
	}
	return nil
}
