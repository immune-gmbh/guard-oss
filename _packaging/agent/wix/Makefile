# command names
CANDLE:=candle.exe
LIGHT:=light.exe

WIXLIBDIR:=$(WIX)bin

# localizations (must match values accepted by light.exe's -cultures switch and requires matching .wxl files)
LOCS:=en-us

# build artifacts, sources
SRCS:=immuneGuard.wixObj wixUi_agent.wixObj enrollKeyDlg.wixObj licenseDlg_immune.wixObj invalidEnrollKeyDlg.wixObj 
MORESRCS:=guard-win.exe img/background.bmp img/banner-top.bmp img/immune.ico
TARGETS:=$(foreach loc,$(LOCS),immuneGuard_$(loc).msi)

LIGHTEXTS:=-ext WixUIExtension -ext WixUtilExtension -ext WixDifxAppExtension
CANDLEEXTS:=-ext WixDifxAppExtension
WIXLIBS:="$(WIXLIBDIR)difxapp_x64.wixlib"

# suppress lots of legacy SCCS and RCS lookups
MAKEFLAGS += --no-builtin-rules

.DEFAULT_GOAL:=all
.PHONY: all
all: $(TARGETS)

immuneGuard_%.msi: lang_%.wxl $(SRCS) $(MORESRCS) eula/eula-%.rtf
	$(LIGHT) $(LIGHTEXTS) -sval -cultures:$* -loc $< -out $@ $(SRCS) $(WIXLIBS)

%.wixObj: %.wxs
	$(CANDLE) $(CANDLEEXTS) -o $@ $<

.PHONY: clean
clean:
	rm -f $(TARGETS) $(SRCS)

# disable many builtin rules
.SUFFIXES: