all: main.pdf

main.pdf: main.tex graphics/chapter-bg.png graphics/immune-logo-schrift.eps Makefile graphics/title-bg.png bibtex.bib
	latexmk -pdfxe main.tex
