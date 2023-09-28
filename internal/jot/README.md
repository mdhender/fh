# JOT

## Standard Fields
Copied from https://en.wikipedia.org/wiki/JSON_Web_Token.

### Field types

* `NumericDate`: an integer representing seconds past 1970-01-01 00:00:00Z.

### Commonly-used header fields
The following fields are commonly used in the header of a JWT

#### alg
_Message authentication code algorithm_

The issuer can freely set an algorithm to verify the signature on the token.
However, some supported algorithms are insecure.

**Required**

#### crit
_Critical_

A list of headers that must be understood by the server in order to accept the token as valid.

**Not implemented**

#### cty
_Content type_

If nested signing or encryption is employed, it is recommended to set this to JWT; otherwise, omit this field.

**Not implemented**

#### kid
_Key ID_

A hint indicating which key the client used to generate the token signature.
The server will match this value to a key on file in order to verify that the signature is valid and the token is authentic.

**Required**

#### typ
_Token type_

If present, it must be set to a registered IANA Media Type.

**Must always be `JOT`**

#### x5c
_x.509 Certificate Chain_

A certificate chain in RFC4945 format corresponding to the private key used to generate the token signature.
The server will use this information to verify that the signature is valid and the token is authentic.

**Not implemented**

#### x5u
_x.509 Certificate Chain URL_

A URL where the server can retrieve a certificate chain corresponding to the private key used to generate the token signature.
The server will retrieve and use this information to verify that the signature is authentic.

**Not implemented**

### Standard claim fields
The internet drafts define the following standard fields ("claims") that can be used inside a JWT claim set. 

#### aud
_Audience_

Identifies the recipients that the JWT is intended for.
Each principal intended to process the JWT must identify itself with a value in the audience claim.
If the principal processing the claim does not identify itself with a value in the aud claim when this claim is present, then the JWT must be rejected.

**Not implemented**

#### exp
_Expiration Time_

Identifies the expiration time on and after which the JWT must not be accepted for processing.
The value must be a NumericDate.

**Requiredr**

#### iat
_Issued at_

Identifies the time at which the JWT was issued.
The value must be a NumericDate.

**Not implemented**

#### iss
_Issuer_

Identifies principal that issued the JWT.

**Required**

#### jti
_JWT ID_

Case-sensitive unique identifier of the token even among different issuers.

**Not implemented**

#### nbf
_Not Before_

Identifies the time on which the JWT will start to be accepted for processing.
The value must be a NumericDate.

**Not implemented**

#### sub
_Subject_

Identifies the subject of the JWT.

### Additional claim fields

#### roles
_Assigned roles_

An array of strings, each string being the name of an assigned role.
For example:

    "roles": ["user", "admin"]
