/*eslint-disable block-scoped-var, id-length, no-control-regex, no-magic-numbers, no-prototype-builtins, no-redeclare, no-shadow, no-var, sort-vars*/
(function($protobuf) {
    "use strict";

    // Common aliases
    var $Reader = $protobuf.Reader, $Writer = $protobuf.Writer, $util = $protobuf.util;
    
    // Exported root namespace
    var $root = $protobuf.roots["default"] || ($protobuf.roots["default"] = {});
    
    $root.test_executor = (function() {
    
        /**
         * Namespace test_executor.
         * @exports test_executor
         * @namespace
         */
        var test_executor = {};
    
        /**
         * TestCaseResult enum.
         * @name test_executor.TestCaseResult
         * @enum {number}
         * @property {number} ACCEPTED=0 ACCEPTED value
         * @property {number} REJECTED=1 REJECTED value
         * @property {number} SKIPPED=2 SKIPPED value
         */
        test_executor.TestCaseResult = (function() {
            var valuesById = {}, values = Object.create(valuesById);
            values[valuesById[0] = "ACCEPTED"] = 0;
            values[valuesById[1] = "REJECTED"] = 1;
            values[valuesById[2] = "SKIPPED"] = 2;
            return values;
        })();
    
        test_executor.SuiteTestResults = (function() {
    
            /**
             * Properties of a SuiteTestResults.
             * @memberof test_executor
             * @interface ISuiteTestResults
             * @property {Array.<number>|null} [supportedFeatures] SuiteTestResults supportedFeatures
             * @property {Array.<number>|null} [unsupportedFeatures] SuiteTestResults unsupportedFeatures
             * @property {Array.<test_executor.TestCaseResult>|null} [testCaseResults] SuiteTestResults testCaseResults
             */
    
            /**
             * Constructs a new SuiteTestResults.
             * @memberof test_executor
             * @classdesc Represents a SuiteTestResults.
             * @implements ISuiteTestResults
             * @constructor
             * @param {test_executor.ISuiteTestResults=} [properties] Properties to set
             */
            function SuiteTestResults(properties) {
                this.supportedFeatures = [];
                this.unsupportedFeatures = [];
                this.testCaseResults = [];
                if (properties)
                    for (var keys = Object.keys(properties), i = 0; i < keys.length; ++i)
                        if (properties[keys[i]] != null)
                            this[keys[i]] = properties[keys[i]];
            }
    
            /**
             * SuiteTestResults supportedFeatures.
             * @member {Array.<number>} supportedFeatures
             * @memberof test_executor.SuiteTestResults
             * @instance
             */
            SuiteTestResults.prototype.supportedFeatures = $util.emptyArray;
    
            /**
             * SuiteTestResults unsupportedFeatures.
             * @member {Array.<number>} unsupportedFeatures
             * @memberof test_executor.SuiteTestResults
             * @instance
             */
            SuiteTestResults.prototype.unsupportedFeatures = $util.emptyArray;
    
            /**
             * SuiteTestResults testCaseResults.
             * @member {Array.<test_executor.TestCaseResult>} testCaseResults
             * @memberof test_executor.SuiteTestResults
             * @instance
             */
            SuiteTestResults.prototype.testCaseResults = $util.emptyArray;
    
            /**
             * Creates a new SuiteTestResults instance using the specified properties.
             * @function create
             * @memberof test_executor.SuiteTestResults
             * @static
             * @param {test_executor.ISuiteTestResults=} [properties] Properties to set
             * @returns {test_executor.SuiteTestResults} SuiteTestResults instance
             */
            SuiteTestResults.create = function create(properties) {
                return new SuiteTestResults(properties);
            };
    
            /**
             * Encodes the specified SuiteTestResults message. Does not implicitly {@link test_executor.SuiteTestResults.verify|verify} messages.
             * @function encode
             * @memberof test_executor.SuiteTestResults
             * @static
             * @param {test_executor.ISuiteTestResults} message SuiteTestResults message or plain object to encode
             * @param {$protobuf.Writer} [writer] Writer to encode to
             * @returns {$protobuf.Writer} Writer
             */
            SuiteTestResults.encode = function encode(message, writer) {
                if (!writer)
                    writer = $Writer.create();
                if (message.supportedFeatures != null && message.supportedFeatures.length) {
                    writer.uint32(/* id 1, wireType 2 =*/10).fork();
                    for (var i = 0; i < message.supportedFeatures.length; ++i)
                        writer.int32(message.supportedFeatures[i]);
                    writer.ldelim();
                }
                if (message.unsupportedFeatures != null && message.unsupportedFeatures.length) {
                    writer.uint32(/* id 2, wireType 2 =*/18).fork();
                    for (var i = 0; i < message.unsupportedFeatures.length; ++i)
                        writer.int32(message.unsupportedFeatures[i]);
                    writer.ldelim();
                }
                if (message.testCaseResults != null && message.testCaseResults.length) {
                    writer.uint32(/* id 3, wireType 2 =*/26).fork();
                    for (var i = 0; i < message.testCaseResults.length; ++i)
                        writer.int32(message.testCaseResults[i]);
                    writer.ldelim();
                }
                return writer;
            };
    
            /**
             * Encodes the specified SuiteTestResults message, length delimited. Does not implicitly {@link test_executor.SuiteTestResults.verify|verify} messages.
             * @function encodeDelimited
             * @memberof test_executor.SuiteTestResults
             * @static
             * @param {test_executor.ISuiteTestResults} message SuiteTestResults message or plain object to encode
             * @param {$protobuf.Writer} [writer] Writer to encode to
             * @returns {$protobuf.Writer} Writer
             */
            SuiteTestResults.encodeDelimited = function encodeDelimited(message, writer) {
                return this.encode(message, writer).ldelim();
            };
    
            /**
             * Decodes a SuiteTestResults message from the specified reader or buffer.
             * @function decode
             * @memberof test_executor.SuiteTestResults
             * @static
             * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
             * @param {number} [length] Message length if known beforehand
             * @returns {test_executor.SuiteTestResults} SuiteTestResults
             * @throws {Error} If the payload is not a reader or valid buffer
             * @throws {$protobuf.util.ProtocolError} If required fields are missing
             */
            SuiteTestResults.decode = function decode(reader, length) {
                if (!(reader instanceof $Reader))
                    reader = $Reader.create(reader);
                var end = length === undefined ? reader.len : reader.pos + length, message = new $root.test_executor.SuiteTestResults();
                while (reader.pos < end) {
                    var tag = reader.uint32();
                    switch (tag >>> 3) {
                    case 1:
                        if (!(message.supportedFeatures && message.supportedFeatures.length))
                            message.supportedFeatures = [];
                        if ((tag & 7) === 2) {
                            var end2 = reader.uint32() + reader.pos;
                            while (reader.pos < end2)
                                message.supportedFeatures.push(reader.int32());
                        } else
                            message.supportedFeatures.push(reader.int32());
                        break;
                    case 2:
                        if (!(message.unsupportedFeatures && message.unsupportedFeatures.length))
                            message.unsupportedFeatures = [];
                        if ((tag & 7) === 2) {
                            var end2 = reader.uint32() + reader.pos;
                            while (reader.pos < end2)
                                message.unsupportedFeatures.push(reader.int32());
                        } else
                            message.unsupportedFeatures.push(reader.int32());
                        break;
                    case 3:
                        if (!(message.testCaseResults && message.testCaseResults.length))
                            message.testCaseResults = [];
                        if ((tag & 7) === 2) {
                            var end2 = reader.uint32() + reader.pos;
                            while (reader.pos < end2)
                                message.testCaseResults.push(reader.int32());
                        } else
                            message.testCaseResults.push(reader.int32());
                        break;
                    default:
                        reader.skipType(tag & 7);
                        break;
                    }
                }
                return message;
            };
    
            /**
             * Decodes a SuiteTestResults message from the specified reader or buffer, length delimited.
             * @function decodeDelimited
             * @memberof test_executor.SuiteTestResults
             * @static
             * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
             * @returns {test_executor.SuiteTestResults} SuiteTestResults
             * @throws {Error} If the payload is not a reader or valid buffer
             * @throws {$protobuf.util.ProtocolError} If required fields are missing
             */
            SuiteTestResults.decodeDelimited = function decodeDelimited(reader) {
                if (!(reader instanceof $Reader))
                    reader = new $Reader(reader);
                return this.decode(reader, reader.uint32());
            };
    
            /**
             * Verifies a SuiteTestResults message.
             * @function verify
             * @memberof test_executor.SuiteTestResults
             * @static
             * @param {Object.<string,*>} message Plain object to verify
             * @returns {string|null} `null` if valid, otherwise the reason why it is not
             */
            SuiteTestResults.verify = function verify(message) {
                if (typeof message !== "object" || message === null)
                    return "object expected";
                if (message.supportedFeatures != null && message.hasOwnProperty("supportedFeatures")) {
                    if (!Array.isArray(message.supportedFeatures))
                        return "supportedFeatures: array expected";
                    for (var i = 0; i < message.supportedFeatures.length; ++i)
                        if (!$util.isInteger(message.supportedFeatures[i]))
                            return "supportedFeatures: integer[] expected";
                }
                if (message.unsupportedFeatures != null && message.hasOwnProperty("unsupportedFeatures")) {
                    if (!Array.isArray(message.unsupportedFeatures))
                        return "unsupportedFeatures: array expected";
                    for (var i = 0; i < message.unsupportedFeatures.length; ++i)
                        if (!$util.isInteger(message.unsupportedFeatures[i]))
                            return "unsupportedFeatures: integer[] expected";
                }
                if (message.testCaseResults != null && message.hasOwnProperty("testCaseResults")) {
                    if (!Array.isArray(message.testCaseResults))
                        return "testCaseResults: array expected";
                    for (var i = 0; i < message.testCaseResults.length; ++i)
                        switch (message.testCaseResults[i]) {
                        default:
                            return "testCaseResults: enum value[] expected";
                        case 0:
                        case 1:
                        case 2:
                            break;
                        }
                }
                return null;
            };
    
            /**
             * Creates a SuiteTestResults message from a plain object. Also converts values to their respective internal types.
             * @function fromObject
             * @memberof test_executor.SuiteTestResults
             * @static
             * @param {Object.<string,*>} object Plain object
             * @returns {test_executor.SuiteTestResults} SuiteTestResults
             */
            SuiteTestResults.fromObject = function fromObject(object) {
                if (object instanceof $root.test_executor.SuiteTestResults)
                    return object;
                var message = new $root.test_executor.SuiteTestResults();
                if (object.supportedFeatures) {
                    if (!Array.isArray(object.supportedFeatures))
                        throw TypeError(".test_executor.SuiteTestResults.supportedFeatures: array expected");
                    message.supportedFeatures = [];
                    for (var i = 0; i < object.supportedFeatures.length; ++i)
                        message.supportedFeatures[i] = object.supportedFeatures[i] | 0;
                }
                if (object.unsupportedFeatures) {
                    if (!Array.isArray(object.unsupportedFeatures))
                        throw TypeError(".test_executor.SuiteTestResults.unsupportedFeatures: array expected");
                    message.unsupportedFeatures = [];
                    for (var i = 0; i < object.unsupportedFeatures.length; ++i)
                        message.unsupportedFeatures[i] = object.unsupportedFeatures[i] | 0;
                }
                if (object.testCaseResults) {
                    if (!Array.isArray(object.testCaseResults))
                        throw TypeError(".test_executor.SuiteTestResults.testCaseResults: array expected");
                    message.testCaseResults = [];
                    for (var i = 0; i < object.testCaseResults.length; ++i)
                        switch (object.testCaseResults[i]) {
                        default:
                        case "ACCEPTED":
                        case 0:
                            message.testCaseResults[i] = 0;
                            break;
                        case "REJECTED":
                        case 1:
                            message.testCaseResults[i] = 1;
                            break;
                        case "SKIPPED":
                        case 2:
                            message.testCaseResults[i] = 2;
                            break;
                        }
                }
                return message;
            };
    
            /**
             * Creates a plain object from a SuiteTestResults message. Also converts values to other types if specified.
             * @function toObject
             * @memberof test_executor.SuiteTestResults
             * @static
             * @param {test_executor.SuiteTestResults} message SuiteTestResults
             * @param {$protobuf.IConversionOptions} [options] Conversion options
             * @returns {Object.<string,*>} Plain object
             */
            SuiteTestResults.toObject = function toObject(message, options) {
                if (!options)
                    options = {};
                var object = {};
                if (options.arrays || options.defaults) {
                    object.supportedFeatures = [];
                    object.unsupportedFeatures = [];
                    object.testCaseResults = [];
                }
                if (message.supportedFeatures && message.supportedFeatures.length) {
                    object.supportedFeatures = [];
                    for (var j = 0; j < message.supportedFeatures.length; ++j)
                        object.supportedFeatures[j] = message.supportedFeatures[j];
                }
                if (message.unsupportedFeatures && message.unsupportedFeatures.length) {
                    object.unsupportedFeatures = [];
                    for (var j = 0; j < message.unsupportedFeatures.length; ++j)
                        object.unsupportedFeatures[j] = message.unsupportedFeatures[j];
                }
                if (message.testCaseResults && message.testCaseResults.length) {
                    object.testCaseResults = [];
                    for (var j = 0; j < message.testCaseResults.length; ++j)
                        object.testCaseResults[j] = options.enums === String ? $root.test_executor.TestCaseResult[message.testCaseResults[j]] : message.testCaseResults[j];
                }
                return object;
            };
    
            /**
             * Converts this SuiteTestResults to JSON.
             * @function toJSON
             * @memberof test_executor.SuiteTestResults
             * @instance
             * @returns {Object.<string,*>} JSON object
             */
            SuiteTestResults.prototype.toJSON = function toJSON() {
                return this.constructor.toObject(this, $protobuf.util.toJSONOptions);
            };
    
            return SuiteTestResults;
        })();
    
        return test_executor;
    })();

    return $root;
})(protobuf);
