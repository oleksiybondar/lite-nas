grammar ZpoolStatus;

/*
 * V1 grammar goal:
 * - Parse one or more `zpool status` pool blocks.
 * - Capture metadata lines and config table rows.
 * - Leave semantic validation (column expectations, numeric checks,
 *   vdev typing, etc.) to AST mapping/evaluator layers.
 */

document
    : leadingBlankLines? poolBlock+ EOF
    ;

leadingBlankLines
    : NL+
    ;

poolBlock
    : poolLine
      metadataSection?
      configSection
      NL*
      errorsLine
      trailingBlankLines?
    ;

trailingBlankLines
    : NL+
    ;

metadataSection
    : metadataLine+
    ;

metadataLine
    : stateLine
    | scanLine
    | statusLine
    | actionLine
    | seeLine
    ;

poolLine
    : POOL_KV textLine
    ;

stateLine
    : STATE_KV textLine
    ;

scanLine
    : SCAN_KV textLine
    ;

statusLine
    : STATUS_KV textLine
    ;

actionLine
    : ACTION_KV textLine
    ;

seeLine
    : SEE_KV textLine
    ;

configSection
    : CONFIG_KV NL
      NL*
      configHeaderLine
      configRowLine+
    ;

configHeaderLine
    : headerAtom+ NL
    ;

configRowLine
    : rowAtom+ NL
    ;

errorsLine
    : ERRORS_KV textLine
    ;

textLine
    : textAtom* NL
    ;

textAtom
    : ATOM
    ;

headerAtom
    : ATOM
    ;

rowAtom
    : ATOM
    ;

POOL_KV   : 'pool:';
STATE_KV  : 'state:';
SCAN_KV   : 'scan:';
CONFIG_KV : 'config:';
ERRORS_KV : 'errors:';
STATUS_KV : 'status:';
ACTION_KV : 'action:';
SEE_KV    : 'see:';

ATOM
    : ~[ \t\r\n]+
    ;

WS
    : [ \t]+ -> skip
    ;

NL
    : '\r'? '\n'
    ;
