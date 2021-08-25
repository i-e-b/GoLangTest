module goLangTestEntry

go 1.16

// Not having these equal can trick you!
//      module name      directory
replace goLangTest => ../InitialJunk
replace games => ../games

require games v0.0.0
require goLangTest v0.0.0
