package template

import (
	"fmt"
	"os"
)

/**
 * List - List all the locally available templates.
 *
 * Parameters:
 * None
 *
 * Returns:
 * A wrapped error if any is raised.
 */
func List() error {
	// Check if any template exists locally, if yes, list them to STDOUT
	if templates, err := os.ReadDir(templateDir); err != nil {
		return fmt.Errorf(
			"failed to read the contents of %s directory: %w",
			templateDir,
			err,
		)
	} else if len(templates) != 0 {
		fmt.Printf("Available Templates:\n")
		for _, template := range templates {
			if template.IsDir() {
				fmt.Printf("%s\n", template.Name())
			}
		}
	}

	return nil
}
