BrowserDemo
===========

This is the browser demonstration page

## Browser Setup

Testing Marionette is best done through the Firefox browser.  If you do not have a copy of Firefox, download it [here](https://www.mozilla.org/en-US/firefox/new/).

### Activate the Proxy

Go to:

``Firefox > Preferences > General > Network Proxy``

- Set the proxy button to Manual Proxy Configuration.
- Set the SOCKS host to the machine to the incoming port on the Marionette client (Probably localhost and port 8079)
- Make sure that the SOCKS v5 Radio button is depressed.
- Check the box marked "Proxy DNS when using SOCKS v5"

### Secure the DNS

Although the code can work through the proxy with the above data, Firefox does not yet have its DNS fully going through the proxy.  To fix this:

- Type about:config into the search bar.  This will open the advanced settings for the browser.
- Go to the term media.peerconnection.enabled 
- Set it to false by double clicking on it.

### Activate Marionette

Start Marionette server as:

``marionette server -format http_simple_blocking -socks5``

Start Marionette client as:

``marionette client -format http_simple_blocking``

Start wireshark on the loopback network and watch the packets.

(Note, if wireshark is not displaying the packets as HTTP, go to:

``WireShark > Preferences > Protocols > HTTP``
 
 and add port 8081 to the port list.

### Surf

Look at your favorite webpage.  Note, there are still some issues regarding maintaining the connection that we are working through.  If the connection drops, then:

- Stop the server and the client
- Restart the server and the client (in order)
- Refresh the page
