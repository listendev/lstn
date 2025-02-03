#!/usr/bin/ruby

# Enable all rules by default.
all

# Extend line length, since each sentence should be on a separate line.
rule 'MD013', :line_length => 99999, :ignore_code_blocks => true

# Allow inline html.
exclude_rule 'MD033'

# Allow multiple headers of the same name/value.
exclude_rule 'MD024'

# Allow code blocks to have no language.
exclude_rule 'MD040'

# Allow header levels to increment by multiple levels.
exclude_rule 'MD001'

# Allow for differences in indentation for lists.
exclude_rule 'MD007'

# Allow trailing spaces.
exclude_rule 'MD009'

# Allow custom table formats.
exclude_rule 'MD055'
exclude_rule 'MD057'

# Allow for incrementing numbers for ordered list prefix.
exclude_rule 'MD029'

# Allow for first line in file to be something other than top level header.
exclude_rule 'MD041'
exclude_rule 'MD002'

# Allow for multiple blank lines.
exclude_rule 'MD012'

# Allow for things to not be surrounded by blank lines.
exclude_rule 'MD022'
exclude_rule 'MD032'

# Allow for hard tabs.
exclude_rule 'MD010'

# Allow bare URLs.
exclude_rule 'MD034'

# Allow files to end with something other than newline.
exclude_rule 'MD047'
