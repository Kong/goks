#!/bin/bash
rm -rf lua-tree
luarocks install kong 2.8.0-0 --tree lua-tree
patch -p1 < patches/lua-tree.patch
