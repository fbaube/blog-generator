# bloggen.yml
# Important notes: 
# 1) The "bloggenator" executable loads this file from its CWD 
#    (current working directory), and therefore "bloggenator" 
#    must be run from the directory that this file is in. This 
#    more or less guarantees that relative file URLs in this 
#    file will be relative to the directory of this file, no
#    matter what the processing. 
# 2) YAML itself is a mess, so this file is simplified by using
#    only one (top) level of properties and only a single value 
#    for each property.
# 3) For any property that actually needs to have multiple values, 
#    (a) enclose all of the values within a pair of single quotes, 
#    (b) separate the values by spaces, and (c) perform processing
#    as described next in (4). 
# 4) Based on (2) above, these properties can all be read into 
#    a map[string]string - at least until a point in processing
#    where (3) above applies - i.e. where a property with multiple
#    values has to be unpacked with strings.Split(propstring," ").
# 5) This architecture makes it dead simple to add new properties.
#    For example, one could add properties like repo_protocol,
#    repo_user, repo_pass. 
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
    author: 'EuroDerf'
    frontpageposts: '10'
statics:
    files: 'favicon.ico robots.txt about.png'
    templates: 'about.html'
