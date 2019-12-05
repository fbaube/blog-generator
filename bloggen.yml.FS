# bloggen.yml
# Important notes: 
# 1) The "bloggenator" executable loads this file from its CWD 
#    (current working directory), and therefore "bloggenator" 
#    must be run from the directory that this file is in. This 
#    constrains relative file URLs in this file to be relative
#    to the directory of this file.
# 2) YAML itself is a mess, so this file keeps it simple by using
#    only one (top) level of properties and only a single value 
#    for each property. For a property that actually need to have 
#    multiple values, enclose all of the values within a pair of 
#    single quotes and separate the values by spaces.
# 3) Based on (2) above, these properties can all be read into a 
#    map[string]string - at least until any property with multiple 
#    values has to be unpacked with strings.Split(propstring," ")
#    - which greatly simplifies processing, and makes it simple to
#    add new properties.
#    
folders:
    repo: 'bsrc'
    tmp: 'workdir'
    dest: 'www'
blog:
    url: 'https://www.infojunkie.eu/REDI'
    language: 'en-us'
    description: 'A blog in Go+MD'
    dateformat: '2006.01.02'
    title: 'infoʤunkie'
    author: 'euroerf'
    frontpageposts: '10'
statics:
    files: 'favicon.ico robots.txt about.png'
    templates: 'about.html'
