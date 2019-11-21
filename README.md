# Xml Splitter

A small utility to help with splitting large xml files into smaller files. Utilises Regex to facilitate cleanup and splitting

## Usage

```bash
xml-splitter -in in/ -out out/ -strip "</{0,1}PMCSet>"


Usage of ./xml-splitter:
  -buffer int
        max number of files to hold in buffer before writing (default 20)
  -depth int
        the nesting depth at which to split the XML (default 1)
  -files int
        number of files to process concurrently (default 1)
  -in string
        the folder to process (glob)
  -out string
        the folder output to
  -skip string
        regex for lines that should be skipped (default "(<\\?xml)|(<!DOCTYPE)")
  -strip string
        regex of values to main from lines
```

## License

Copyright (c) 2019, Medicines Discovery Catapult
All rights reserved.

Redistribution and use in source and binary forms, with or without
modification, are permitted provided that the following conditions are met:
    * Redistributions of source code must retain the above copyright
      notice, this list of conditions and the following disclaimer.
    * Redistributions in binary form must reproduce the above copyright
      notice, this list of conditions and the following disclaimer in the
      documentation and/or other materials provided with the distribution.
    * Neither the name of the Medicines Discovery Catapult nor the
      names of its contributors may be used to endorse or promote products
      derived from this software without specific prior written permission.

THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS "AS IS" AND
ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT LIMITED TO, THE IMPLIED
WARRANTIES OF MERCHANTABILITY AND FITNESS FOR A PARTICULAR PURPOSE ARE
DISCLAIMED. IN NO EVENT SHALL MEDICINES DISCOVERY CATAPULT BE LIABLE FOR ANY
DIRECT, INDIRECT, INCIDENTAL, SPECIAL, EXEMPLARY, OR CONSEQUENTIAL DAMAGES
(INCLUDING, BUT NOT LIMITED TO, PROCUREMENT OF SUBSTITUTE GOODS OR SERVICES;
LOSS OF USE, DATA, OR PROFITS; OR BUSINESS INTERRUPTION) HOWEVER CAUSED AND
ON ANY THEORY OF LIABILITY, WHETHER IN CONTRACT, STRICT LIABILITY, OR TORT
(INCLUDING NEGLIGENCE OR OTHERWISE) ARISING IN ANY WAY OUT OF THE USE OF THIS
SOFTWARE, EVEN IF ADVISED OF THE POSSIBILITY OF SUCH DAMAGE.