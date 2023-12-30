immune Guard Technology Whitepaper
==================================

You can get the latest PDF build from the 'Actions' tab. From there select the
latest build and download the 'document' artifact.

Build
-----

The LaTeX document is built using XeTeX. Aside from XeTeX you'll need memoir,
TikZ and the Fira Code, Asap and Eurostile fonts.

**Gentoo**

```bash
USE="xetex graphics" sudo -E emerge -av texlive latexmk
cp fonts/* ~/.local/share/fonts/
```

**Ubuntu**

```bash
sudo apt install texlive-xetex latexmk texlive-science
mkdir ~/.fonts
cp fonts/* ~/.fonts/
```

After all dependencies are installed build the PDF using make:

```bash
make
```

Then, open `main.pdf`.
