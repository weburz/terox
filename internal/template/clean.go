package template

import (
	"fmt"
	"os"
	"path/filepath"
)

/**
 * Clean: Cleanup the locally downloaded templates.
 *
 * Parameters:
 *   None
 *
 * Returns:
 *   An error if any was raised during the removal process.
 */
func Clean() error {
	// Read the contents of the template directory to check for templates
	if templates, err := os.ReadDir(templateDir); err != nil {
		return fmt.Errorf(
			"failed to find any templates at %s: %w",
			templateDir,
			err,
		)
	} else if len(templates) != 0 {
		fmt.Printf("The following templates were deleted:\n\n")
		for _, template := range templates {
			path := filepath.Join(templateDir, template.Name())
			fmt.Printf("%s\n", template.Name())
			if err := os.RemoveAll(path); err != nil {
				return fmt.Errorf(
					"failed to remove %s: %w",
					template.Name(),
					err,
				)
			}
		}
	}

	return nil
}
