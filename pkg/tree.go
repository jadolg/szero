package pkg

import (
	"fmt"
	"io"
	"os"

	"github.com/charmbracelet/lipgloss"
)

// TreePrinter provides a tree-style output for scaling operations
type TreePrinter struct {
	writer io.Writer
}

// ScaleInfo contains information about a scaling operation
type ScaleInfo struct {
	Name     string
	Replicas int32
	Scaled   bool
	Warning  string // if not scaled, this contains the reason
}

// ResourceGroup groups resources by type for tree output
type ResourceGroup struct {
	Type      string // "Deployments", "StatefulSets", "DaemonSets"
	Resources []ScaleInfo
	Skipped   bool
}

// NamespaceResult contains all scaling results for a namespace
type NamespaceResult struct {
	Namespace    string
	Deployments  ResourceGroup
	StatefulSets ResourceGroup
	DaemonSets   ResourceGroup
}

var (
	namespaceStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("212"))
	resourceStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("39"))
	itemStyle      = lipgloss.NewStyle() // Use default terminal color for visibility on both themes
	replicaStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("212")).Bold(true)
	warnStyle      = lipgloss.NewStyle().Foreground(lipgloss.Color("214"))
	skipStyle      = lipgloss.NewStyle().Italic(true) // Use default color with italic for visibility
)

// NewTreePrinter creates a new TreePrinter
func NewTreePrinter() *TreePrinter {
	return &TreePrinter{writer: os.Stdout}
}

// NewTreePrinterWithWriter creates a TreePrinter with a custom writer
func NewTreePrinterWithWriter(w io.Writer) *TreePrinter {
	return &TreePrinter{writer: w}
}

// PrintNamespaceResult prints the scaling result for a namespace in tree format
func (tp *TreePrinter) PrintNamespaceResult(result NamespaceResult) error {
	// Print namespace header
	if _, err := fmt.Fprintf(tp.writer, "%s\n", namespaceStyle.Render(result.Namespace)); err != nil {
		return err
	}

	groups := []ResourceGroup{result.Deployments, result.StatefulSets, result.DaemonSets}

	for i, group := range groups {
		isLast := i == len(groups)-1
		if err := tp.printResourceGroup(group, isLast); err != nil {
			return err
		}
	}
	_, err := fmt.Fprintln(tp.writer)
	return err
}

func (tp *TreePrinter) printResourceGroup(group ResourceGroup, isLast bool) error {
	connector := "├── "
	childPrefix := "│   "
	if isLast {
		connector = "└── "
		childPrefix = "    "
	}

	if group.Skipped {
		_, err := fmt.Fprintf(tp.writer, "%s%s\n", connector, skipStyle.Render(fmt.Sprintf("%s (skipped)", group.Type)))
		return err
	}

	// Print resource type header with count
	scaledCount := 0
	for _, r := range group.Resources {
		if r.Scaled {
			scaledCount++
		}
	}
	header := fmt.Sprintf("%s (%d/%d)", group.Type, scaledCount, len(group.Resources))
	if _, err := fmt.Fprintf(tp.writer, "%s%s\n", connector, resourceStyle.Render(header)); err != nil {
		return err
	}

	// Print each resource
	for j, res := range group.Resources {
		isLastItem := j == len(group.Resources)-1
		itemConnector := "├── "
		if isLastItem {
			itemConnector = "└── "
		}

		if res.Scaled {
			var info string
			if res.Replicas > 0 {
				info = fmt.Sprintf("%s → %s", res.Name, replicaStyle.Render(fmt.Sprintf("%d replicas", res.Replicas)))
			} else {
				info = res.Name
			}
			if _, err := fmt.Fprintf(tp.writer, "%s%s%s\n", childPrefix, itemConnector, itemStyle.Render(info)); err != nil {
				return err
			}
		} else {
			info := fmt.Sprintf("%s %s", res.Name, warnStyle.Render(fmt.Sprintf("(%s)", res.Warning)))
			if _, err := fmt.Fprintf(tp.writer, "%s%s%s\n", childPrefix, itemConnector, itemStyle.Render(info)); err != nil {
				return err
			}
		}
	}
	return nil
}
