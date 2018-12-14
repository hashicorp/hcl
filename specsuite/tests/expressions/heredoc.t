result = {
    normal = {
        basic         = "Foo\nBar\nBaz\n"
        indented      = "    Foo\n    Bar\n    Baz\n"
        indented_more = "    Foo\n      Bar\n    Baz\n"
        interp        = "    Foo\n    Bar\n    Baz\n"

        marker_at_suffix = "    NOT EOT\n"
    }
    flush  = {
        basic                = "Foo\nBar\nBaz\n"
        indented             = "Foo\nBar\nBaz\n"
        indented_more        = "Foo\n  Bar\nBaz\n"
        indented_less        = "  Foo\nBar\n  Baz\n"
        interp               = "Foo\nBar\nBaz\n"
        interp_indented_more = "Foo\n  Bar\nBaz\n"
        interp_indented_less = "  Foo\n  Bar\n  Baz\n"
        tabs                 = "Foo\n Bar\n Baz\n"
        unicode_spaces       = " Foo (there's two \"em spaces\" before Foo there)\nBar\nBaz\n"
    }
}
result_type = object({
  normal = map(string)
  flush  = map(string)
})
