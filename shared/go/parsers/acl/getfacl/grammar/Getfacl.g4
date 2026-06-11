grammar Getfacl;

document
    : line* EOF
    ;

line
    : headerLine NL
    | aclEntryLine NL
    | commentLine NL
    | NL
    ;

headerLine
    : HASH WS* headerKey COLON WS* valueAtom?
    ;

commentLine
    : HASH valueAtom?
    ;

aclEntryLine
    : DEFAULT_PREFIX? tag COLON qualifier COLON PERM
    ;

headerKey
    : 'file'
    | 'owner'
    | 'group'
    ;

tag
    : USER_TAG
    | GROUP_TAG
    | OTHER_TAG
    | MASK_TAG
    ;

qualifier
    : valueAtom?
    ;

valueAtom
    : VALUE_ATOM
    ;

DEFAULT_PREFIX : 'default:';
USER_TAG       : 'user';
GROUP_TAG      : 'group';
OTHER_TAG      : 'other';
MASK_TAG       : 'mask';

PERM : [r-][w-][x-];
VALUE_ATOM : ~[: \t\r\n#]+;

HASH : '#';
COLON : ':';
WS : [ \t]+ -> skip;
NL : '\r'? '\n';
