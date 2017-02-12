## Font For All (FFA)
 FFA is a Windows service to allow users to "install" fonts on Windows without having administrative privileges or write access to %systemroot%\fonts and the registry. The application makes use of the AddFontResource and is therefore needed to be run on boot, otherwise the fonts will be removed from the font-list on reboot. 

### Installation
Unpack the executable to a directory of your choice (preferably a location where you won't delete it by chance such as `c:\Program Files\` or similar). Open a command-prompt with administrator privileges and run

```
C:\Program Files\FontForAll\> fontforall.exe install
C:\Program Files\FontForAll\> sc start fontforall
```
Make sure the service is running by either opening `services.msc` or running `sc query fontforall` in the command-prompt.

### How it works
FFA creates a new directory under `%PUBLIC%\fonts` (normally `C:\Users\Public\fonts`) where all users can put their custom fonts. FFA will watch this directory for changes and install those fonts for all users that sign in.

#### Adding fonts
Just copy/move them into `%PUBLIC%\fonts` (normally `C:\Users\Public\fonts`) and start/restart the application you want to have access to the font.

#### Deleting fonts
Deleting fonts is a bit more complicated since fonts loaded using this method can't be modified or removed without rebooting the system first. FFA will normally load all fonts in the directory on boot, so FFA must first be stopped and disabled through the Service Manager (`services.msc`), reboot the computer and re-enable FFA through the Service Manager again.
