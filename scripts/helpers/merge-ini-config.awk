# Merge a template into an installed INI file while keeping installed values.
# The first input is the template; the second is the installed file.
function fail(message) {
	print "Invalid INI config " FILENAME ":" FNR ": " message > "/dev/stderr"
	failed = 1
	exit 2
}

function parse_line(line,    trimmed, separator, key) {
	trimmed = line
	sub(/^[[:space:]]+/, "", trimmed)
	sub(/[[:space:]]+$/, "", trimmed)
	if (trimmed == "" || trimmed ~ /^[#;]/) {
		return "comment"
	}
	if (trimmed ~ /^\[[^][]+\]$/) {
		section = substr(trimmed, 2, length(trimmed) - 2)
		return "section"
	}
	separator = index(trimmed, "=")
	if (section == "" || separator == 0) {
		fail("expected a section or key=value")
	}
	key = substr(trimmed, 1, separator - 1)
	sub(/[[:space:]]+$/, "", key)
	if (key !~ /^[A-Za-z0-9_.-]+$/) {
		fail("invalid key")
	}
	parsed_key = key
	return "key"
}

function emit_missing(name,    i, key, index_key) {
	for (i = 1; i <= template_key_count[name]; i++) {
		key = template_key[name, i]
		index_key = name SUBSEP key
		if (!(index_key in installed_key)) {
			print template_line[index_key]
		}
	}
}

FNR == 1 {
	section = ""
}

NR == FNR {
	kind = parse_line($0)
	if (kind == "section") {
		if (section in template_section) {
			fail("duplicate section " section)
		}
		template_section[section] = 1
		template_section_order[++template_section_count] = section
	} else if (kind == "key") {
		index_key = section SUBSEP parsed_key
		if (index_key in template_line) {
			fail("duplicate key " parsed_key)
		}
		template_key[section, ++template_key_count[section]] = parsed_key
		template_line[index_key] = $0
	}
	next
}

{
	previous_section = section
	kind = parse_line($0)
	if (kind == "section") {
		if (section in installed_section) {
			fail("duplicate section " section)
		}
		installed_section[section] = 1
		if (previous_section != "") {
			emit_missing(previous_section)
		}
	} else if (kind == "key") {
		index_key = section SUBSEP parsed_key
		if (index_key in installed_key) {
			fail("duplicate key " parsed_key)
		}
		installed_key[index_key] = 1
	}
	print
}

END {
	if (failed) {
		exit 2
	}
	if (section != "") {
		emit_missing(section)
	}
	for (i = 1; i <= template_section_count; i++) {
		name = template_section_order[i]
		if (!(name in installed_section)) {
			print ""
			print "[" name "]"
			emit_missing(name)
		}
	}
}
