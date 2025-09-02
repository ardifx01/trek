package main

import (
	"archive/zip"
	"io"
	"io/fs"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

var rootDir = "../result"

func ZipFolder(source, target string) error {
	zipFile, err := os.Create(target)
	if err != nil {
		return err
	}
	defer zipFile.Close()

	archive := zip.NewWriter(zipFile)
	defer archive.Close()

	err = filepath.Walk(source, func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}

		relativePath, err := filepath.Rel(filepath.Dir(source), path)
		if err != nil {
			return err
		}

		if info.IsDir() {
			_, err = archive.Create(relativePath + "/")
			return err
		}

		file, err := os.Open(path)
		if err != nil {
			return err
		}
		defer file.Close()

		w, err := archive.Create(relativePath)
		if err != nil {
			return err
		}
		_, err = io.Copy(w, file)
		return err
	})
	return err
}

func main() {
	// cek apakah folder result ada atau tidak, jika tidak ada maka create baru
	_, err := os.Stat(rootDir)
	if os.IsNotExist(err) {
		os.Mkdir(rootDir, 0755)
	}

	root := tview.NewTreeNode(rootDir).
		SetColor(tcell.ColorRed)
	tree := tview.NewTreeView().
		SetRoot(root).
		SetCurrentNode(root)
	tree.SetBorder(true)
	tree.SetTitle(`Add "/" at the end to create a folder`).SetTitleAlign(tview.AlignRight)

	add := func(target *tview.TreeNode, path string) {
		files, err := ioutil.ReadDir(path)
		if err != nil {
			panic(err)
		}
		for _, file := range files {
			node := tview.NewTreeNode(file.Name()).
				SetReference(filepath.Join(path, file.Name())).
				SetSelectable(file.IsDir())
			if file.IsDir() {
				node.SetColor(tcell.ColorGreen)
			}
			target.AddChild(node)
		}
	}

	add(root, rootDir)

	tree.SetSelectedFunc(func(node *tview.TreeNode) {
		reference := node.GetReference()
		if reference == nil {
			return
		}
		children := node.GetChildren()
		if len(children) == 0 {
			path := reference.(string)
			add(node, path)
		} else {
			node.SetExpanded(!node.IsExpanded())
		}
	})

	/*
		Di dalam layout grid, ada function SetRows dan SetColumns
		SetRows: berguna untuk mengatur, berapa banyak row yang ingin kita gunakan
		SetColumn: bergunakan untuk mengatur besaran width dari kolom tersebut, semisal
		SetColumn(20, 0), artinya di kolom ke 1 itu akan selalu widthnya 20, sedangkan jika kita pilih kolom ke 2, maka widthnya itu flexble, atau mengambil sisa cell di kananya
	*/
	grid := tview.NewGrid().SetRows(2).SetColumns(80, 0)
	grid.SetBorder(true)
	header := tview.NewTextView().SetText("Trek").SetTextAlign(tview.AlignCenter)

	namaFolderInput := tview.NewInputField().SetLabel("Nama File/Folder:")
	form := tview.NewForm()
	form.SetBorder(true)
	form.AddFormItem(namaFolderInput).AddButton("Create", func() {
		folderName := namaFolderInput.GetText()
		if folderName == "" {
			return
		}

		selectedTreeFolder := tree.GetCurrentNode()
		var basePath string
		if selectedTreeFolder == nil {
			basePath = rootDir
		} else {
			ref := selectedTreeFolder.GetReference()
			if ref == nil {
				basePath = rootDir
			} else {
				basePath = ref.(string)
			}
		}

		path := filepath.Join(basePath, folderName)

		if strings.Contains(folderName, "/") {
			if err := os.Mkdir(path, 0755); err != nil {
				panic(err)
			}
		} else {
			f, err := os.Create(path)
			if err != nil {
				log.Fatal("error while create file:", err)
			}
			defer f.Close()
		}

		selectedTreeFolder.ClearChildren()
		add(selectedTreeFolder, basePath)
		selectedTreeFolder.SetExpanded(true)

		namaFolderInput.SetText("")
	}).AddButton("Save ZIP", func() {
		err := ZipFolder(rootDir, "output.zip")
		if err != nil {
			log.Fatal("failed to zip:", err)
		}
	})

	app := tview.NewApplication()
	aboutBtn := tview.NewButton("About")
	aboutBtn.SetSelectedFunc(func() {
		modal := tview.NewModal().SetText("Trek (Tree + Explorer) created by aji mustofa @pepega90").AddButtons([]string{"Keluar"}).SetDoneFunc(func(buttonIndex int, buttonLabel string) {
			if buttonLabel == "Keluar" {
				app.SetRoot(grid, true)
			}
		})
		app.SetRoot(modal, true)
	})

	// layouts
	flexFormTree := tview.NewFlex().AddItem(form, 50, 0, false).AddItem(tree, 0, 1, false)
	headerFlex := tview.NewFlex().AddItem(aboutBtn, 10, 0, false).AddItem(header, 0, 1, false)
	grid.AddItem(headerFlex, 0, 0, 1, 2, 0, 0, false)
	// grid.AddItem(sideBar, 0, 0, 1, 1, 0, 0, false)
	grid.AddItem(flexFormTree, 1, 0, 1, 2, 0, 0, true)

	// list of views
	views := []tview.Primitive{form, tree}
	currentFocus := 0

	app.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyRight {
			currentFocus = (currentFocus + 1) % len(views)
			app.SetFocus(views[currentFocus])
			return nil
		} else if event.Key() == tcell.KeyLeft {
			currentFocus = (currentFocus + 1) % len(views)
			app.SetFocus(views[currentFocus])
			return nil
		} else if event.Key() == tcell.KeyCtrlR {
			app.SetFocus(aboutBtn)
			return nil
		}
		return event
	})

	if err := app.SetRoot(grid, true).Run(); err != nil {
		panic(err)
	}
}
