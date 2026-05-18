/**
 * SCORM Runtime API Implementation
 * Supports both SCORM 1.2 and SCORM 2004
 */

(function() {
    'use strict';

    // SCORM Runtime Configuration
    var SCORM_RUNTIME = {
        version: '1.2', // Default version
        debug: true,
        api: null,
        data: {},
        initialized: false,
        terminated: false,
        commitPending: false,
        baseUrl: '/api/scorm',
        attemptId: null,
        sessionId: null
    };

    // SCORM 1.2 API Implementation
    var SCORM12_API = {
        Clever SchoolInitialize: function(param) {
            if (SCORM_RUNTIME.debug) console.log('SCORM 1.2: Clever SchoolInitialize called with:', param);
            
            if (SCORM_RUNTIME.initialized) {
                return "false";
            }
            
            SCORM_RUNTIME.initialized = true;
            SCORM_RUNTIME.attemptId = param || generateAttemptId();
            SCORM_RUNTIME.sessionId = generateSessionId();
            
            // Initialize data structure
            SCORM_RUNTIME.data = {
                'cmi.core.lesson_status': 'not attempted',
                'cmi.core.lesson_location': '',
                'cmi.core.score.raw': '',
                'cmi.core.score.min': '',
                'cmi.core.score.max': '',
                'cmi.core.total_time': '0000:00:00',
                'cmi.core.session_time': '0000:00:00',
                'cmi.core.entry': 'resume',
                'cmi.core.exit': '',
                'cmi.core.student_id': '',
                'cmi.core.student_name': '',
                'cmi.core.lesson_mode': 'normal',
                'cmi.core.credit': 'credit',
                'cmi.core.audio': 'unknown',
                'cmi.core.language': 'unknown',
                'cmi.core.audio_level': 'unknown',
                'cmi.core.audio_captioning': 'unknown',
                'cmi.core.delivery_speed': 'unknown',
                'cmi.core.delivery_format': 'unknown',
                'cmi.core.completion_threshold': 'unknown',
                'cmi.core.completion_status': 'unknown',
                'cmi.core.success_status': 'unknown',
                'cmi.core.max_time_allowed': 'unknown',
                'cmi.core.time_limit_action': 'unknown',
                'cmi.core.data_from_Clever School': '',
                'cmi.core.mastery_score': 'unknown',
                'cmi.core.launch_data': '',
                'cmi.core.lesson_mode': 'normal',
                'cmi.core.credit': 'credit',
                'cmi.core.audio': 'unknown',
                'cmi.core.language': 'unknown',
                'cmi.core.audio_level': 'unknown',
                'cmi.core.audio_captioning': 'unknown',
                'cmi.core.delivery_speed': 'unknown',
                'cmi.core.delivery_format': 'unknown',
                'cmi.core.completion_threshold': 'unknown',
                'cmi.core.completion_status': 'unknown',
                'cmi.core.success_status': 'unknown',
                'cmi.core.max_time_allowed': 'unknown',
                'cmi.core.time_limit_action': 'unknown',
                'cmi.core.data_from_Clever School': '',
                'cmi.core.mastery_score': 'unknown',
                'cmi.core.launch_data': ''
            };
            
            return "true";
        },

        Clever SchoolFinish: function(param) {
            if (SCORM_RUNTIME.debug) console.log('SCORM 1.2: Clever SchoolFinish called with:', param);
            
            if (!SCORM_RUNTIME.initialized) {
                return "false";
            }
            
            // Commit any pending data
            if (SCORM_RUNTIME.commitPending) {
                this.Clever SchoolCommit("");
            }
            
            // Terminate the session
            this.Clever SchoolTerminate("");
            
            return "true";
        },

        Clever SchoolGetValue: function(element) {
            if (SCORM_RUNTIME.debug) console.log('SCORM 1.2: Clever SchoolGetValue called for:', element);
            
            if (!SCORM_RUNTIME.initialized) {
                return "";
            }
            
            // Check if element exists in our data
            if (SCORM_RUNTIME.data.hasOwnProperty(element)) {
                return SCORM_RUNTIME.data[element];
            }
            
            // Try to get from server
            return getFromServer(element);
        },

        Clever SchoolSetValue: function(element, value) {
            if (SCORM_RUNTIME.debug) console.log('SCORM 1.2: Clever SchoolSetValue called for:', element, 'with value:', value);
            
            if (!SCORM_RUNTIME.initialized) {
                return "false";
            }
            
            // Validate element and value
            if (!isValidElement(element) || !isValidValue(element, value)) {
                return "false";
            }
            
            // Store locally
            SCORM_RUNTIME.data[element] = value;
            SCORM_RUNTIME.commitPending = true;
            
            // Send to server
            setToServer(element, value);
            
            return "true";
        },

        Clever SchoolCommit: function(param) {
            if (SCORM_RUNTIME.debug) console.log('SCORM 1.2: Clever SchoolCommit called with:', param);
            
            if (!SCORM_RUNTIME.initialized) {
                return "false";
            }
            
            // Commit data to server
            commitToServer();
            SCORM_RUNTIME.commitPending = false;
            
            return "true";
        },

        Clever SchoolGetLastError: function() {
            return SCORM_RUNTIME.lastError || "0";
        },

        Clever SchoolGetErrorString: function(errorCode) {
            var errorMessages = {
                "0": "No error",
                "101": "General exception",
                "201": "Invalid argument error",
                "301": "Element cannot have children",
                "401": "Element not an array - cannot have count",
                "402": "Element not an array - cannot set count",
                "403": "Element not an array - first index is 0",
                "404": "Element not an array - array index must be positive",
                "405": "Element not an array - this is a read-only element",
                "406": "Element not an array - array index is too high",
                "407": "Element not an array - array index is too low",
                "408": "Element not an array - array index is not a number",
                "409": "Element not an array - array index is not an integer",
                "410": "Element not an array - array index is not a positive integer",
                "411": "Element not an array - array index is not a positive integer",
                "412": "Element not an array - array index is not a positive integer",
                "413": "Element not an array - array index is not a positive integer",
                "414": "Element not an array - array index is not a positive integer",
                "415": "Element not an array - array index is not a positive integer",
                "416": "Element not an array - array index is not a positive integer",
                "417": "Element not an array - array index is not a positive integer",
                "418": "Element not an array - array index is not a positive integer",
                "419": "Element not an array - array index is not a positive integer",
                "420": "Element not an array - array index is not a positive integer",
                "421": "Element not an array - array index is not a positive integer",
                "422": "Element not an array - array index is not a positive integer",
                "423": "Element not an array - array index is not a positive integer",
                "424": "Element not an array - array index is not a positive integer",
                "425": "Element not an array - array index is not a positive integer",
                "426": "Element not an array - array index is not a positive integer",
                "427": "Element not an array - array index is not a positive integer",
                "428": "Element not an array - array index is not a positive integer",
                "429": "Element not an array - array index is not a positive integer",
                "430": "Element not an array - array index is not a positive integer",
                "431": "Element not an array - array index is not a positive integer",
                "432": "Element not an array - array index is not a positive integer",
                "433": "Element not an array - array index is not a positive integer",
                "434": "Element not an array - array index is not a positive integer",
                "435": "Element not an array - array index is not a positive integer",
                "436": "Element not an array - array index is not a positive integer",
                "437": "Element not an array - array index is not a positive integer",
                "438": "Element not an array - array index is not a positive integer",
                "439": "Element not an array - array index is not a positive integer",
                "440": "Element not an array - array index is not a positive integer",
                "441": "Element not an array - array index is not a positive integer",
                "442": "Element not an array - array index is not a positive integer",
                "443": "Element not an array - array index is not a positive integer",
                "444": "Element not an array - array index is not a positive integer",
                "445": "Element not an array - array index is not a positive integer",
                "446": "Element not an array - array index is not a positive integer",
                "447": "Element not an array - array index is not a positive integer",
                "448": "Element not an array - array index is not a positive integer",
                "449": "Element not an array - array index is not a positive integer",
                "450": "Element not an array - array index is not a positive integer",
                "451": "Element not an array - array index is not a positive integer",
                "452": "Element not an array - array index is not a positive integer",
                "453": "Element not an array - array index is not a positive integer",
                "454": "Element not an array - array index is not a positive integer",
                "455": "Element not an array - array index is not a positive integer",
                "456": "Element not an array - array index is not a positive integer",
                "457": "Element not an array - array index is not a positive integer",
                "458": "Element not an array - array index is not a positive integer",
                "459": "Element not an array - array index is not a positive integer",
                "460": "Element not an array - array index is not a positive integer",
                "461": "Element not an array - array index is not a positive integer",
                "462": "Element not an array - array index is not a positive integer",
                "463": "Element not an array - array index is not a positive integer",
                "464": "Element not an array - array index is not a positive integer",
                "465": "Element not an array - array index is not a positive integer",
                "466": "Element not an array - array index is not a positive integer",
                "467": "Element not an array - array index is not a positive integer",
                "468": "Element not an array - array index is not a positive integer",
                "469": "Element not an array - array index is not a positive integer",
                "470": "Element not an array - array index is not a positive integer",
                "471": "Element not an array - array index is not a positive integer",
                "472": "Element not an array - array index is not a positive integer",
                "473": "Element not an array - array index is not a positive integer",
                "474": "Element not an array - array index is not a positive integer",
                "475": "Element not an array - array index is not a positive integer",
                "476": "Element not an array - array index is not a positive integer",
                "477": "Element not an array - array index is not a positive integer",
                "478": "Element not an array - array index is not a positive integer",
                "479": "Element not an array - array index is not a positive integer",
                "480": "Element not an array - array index is not a positive integer",
                "481": "Element not an array - array index is not a positive integer",
                "482": "Element not an array - array index is not a positive integer",
                "483": "Element not an array - array index is not a positive integer",
                "484": "Element not an array - array index is not a positive integer",
                "485": "Element not an array - array index is not a positive integer",
                "486": "Element not an array - array index is not a positive integer",
                "487": "Element not an array - array index is not a positive integer",
                "488": "Element not an array - array index is not a positive integer",
                "489": "Element not an array - array index is not a positive integer",
                "490": "Element not an array - array index is not a positive integer",
                "491": "Element not an array - array index is not a positive integer",
                "492": "Element not an array - array index is not a positive integer",
                "493": "Element not an array - array index is not a positive integer",
                "494": "Element not an array - array index is not a positive integer",
                "495": "Element not an array - array index is not a positive integer",
                "496": "Element not an array - array index is not a positive integer",
                "497": "Element not an array - array index is not a positive integer",
                "498": "Element not an array - array index is not a positive integer",
                "499": "Element not an array - array index is not a positive integer",
                "500": "Element not an array - array index is not a positive integer"
            };
            
            return errorMessages[errorCode] || "Unknown error";
        },

        Clever SchoolGetDiagnostic: function(errorCode) {
            return this.Clever SchoolGetErrorString(errorCode);
        },

        Clever SchoolTerminate: function(param) {
            if (SCORM_RUNTIME.debug) console.log('SCORM 1.2: Clever SchoolTerminate called with:', param);
            
            if (!SCORM_RUNTIME.initialized) {
                return "false";
            }
            
            // Commit any pending data
            if (SCORM_RUNTIME.commitPending) {
                this.Clever SchoolCommit("");
            }
            
            // Terminate session on server
            terminateOnServer();
            
            SCORM_RUNTIME.terminated = true;
            SCORM_RUNTIME.initialized = false;
            
            return "true";
        }
    };

    // SCORM 2004 API Implementation
    var SCORM2004_API = {
        Initialize: function(param) {
            if (SCORM_RUNTIME.debug) console.log('SCORM 2004: Initialize called with:', param);
            
            if (SCORM_RUNTIME.initialized) {
                return "false";
            }
            
            SCORM_RUNTIME.initialized = true;
            SCORM_RUNTIME.attemptId = param || generateAttemptId();
            SCORM_RUNTIME.sessionId = generateSessionId();
            
            // Initialize data structure for 2004
            SCORM_RUNTIME.data = {
                'cmi.completion_status': 'not attempted',
                'cmi.success_status': 'unknown',
                'cmi.score.scaled': '',
                'cmi.score.raw': '',
                'cmi.score.min': '',
                'cmi.score.max': '',
                'cmi.total_time': 'PT0H0M0S',
                'cmi.session_time': 'PT0H0M0S',
                'cmi.entry': 'resume',
                'cmi.exit': '',
                'cmi.student_id': '',
                'cmi.student_name': '',
                'cmi.lesson_mode': 'normal',
                'cmi.credit': 'credit',
                'cmi.audio_level': 'unknown',
                'cmi.audio_captioning': 'unknown',
                'cmi.delivery_speed': 'unknown',
                'cmi.delivery_format': 'unknown',
                'cmi.completion_threshold': 'unknown',
                'cmi.max_time_allowed': 'unknown',
                'cmi.time_limit_action': 'unknown',
                'cmi.data_from_Clever School': '',
                'cmi.mastery_score': 'unknown',
                'cmi.launch_data': '',
                'cmi.learner_preference.audio_level': 'unknown',
                'cmi.learner_preference.language': 'unknown',
                'cmi.learner_preference.delivery_speed': 'unknown',
                'cmi.learner_preference.audio_captioning': 'unknown',
                'cmi.learner_preference.color': 'unknown',
                'cmi.learner_preference.content': 'unknown',
                'cmi.learner_preference.audio': 'unknown',
                'cmi.learner_preference.text': 'unknown',
                'cmi.learner_preference.video': 'unknown',
                'cmi.learner_preference.audio_level': 'unknown',
                'cmi.learner_preference.language': 'unknown',
                'cmi.learner_preference.delivery_speed': 'unknown',
                'cmi.learner_preference.audio_captioning': 'unknown',
                'cmi.learner_preference.color': 'unknown',
                'cmi.learner_preference.content': 'unknown',
                'cmi.learner_preference.audio': 'unknown',
                'cmi.learner_preference.text': 'unknown',
                'cmi.learner_preference.video': 'unknown'
            };
            
            return "true";
        },

        Terminate: function(param) {
            if (SCORM_RUNTIME.debug) console.log('SCORM 2004: Terminate called with:', param);
            
            if (!SCORM_RUNTIME.initialized) {
                return "false";
            }
            
            // Commit any pending data
            if (SCORM_RUNTIME.commitPending) {
                this.Commit("");
            }
            
            // Terminate session on server
            terminateOnServer();
            
            SCORM_RUNTIME.terminated = true;
            SCORM_RUNTIME.initialized = false;
            
            return "true";
        },

        GetValue: function(element) {
            if (SCORM_RUNTIME.debug) console.log('SCORM 2004: GetValue called for:', element);
            
            if (!SCORM_RUNTIME.initialized) {
                return "";
            }
            
            // Check if element exists in our data
            if (SCORM_RUNTIME.data.hasOwnProperty(element)) {
                return SCORM_RUNTIME.data[element];
            }
            
            // Try to get from server
            return getFromServer(element);
        },

        SetValue: function(element, value) {
            if (SCORM_RUNTIME.debug) console.log('SCORM 2004: SetValue called for:', element, 'with value:', value);
            
            if (!SCORM_RUNTIME.initialized) {
                return "false";
            }
            
            // Validate element and value
            if (!isValidElement(element) || !isValidValue(element, value)) {
                return "false";
            }
            
            // Store locally
            SCORM_RUNTIME.data[element] = value;
            SCORM_RUNTIME.commitPending = true;
            
            // Send to server
            setToServer(element, value);
            
            return "true";
        },

        Commit: function(param) {
            if (SCORM_RUNTIME.debug) console.log('SCORM 2004: Commit called with:', param);
            
            if (!SCORM_RUNTIME.initialized) {
                return "false";
            }
            
            // Commit data to server
            commitToServer();
            SCORM_RUNTIME.commitPending = false;
            
            return "true";
        },

        GetLastError: function() {
            return SCORM_RUNTIME.lastError || "0";
        },

        GetErrorString: function(errorCode) {
            return SCORM12_API.Clever SchoolGetErrorString(errorCode);
        },

        GetDiagnostic: function(errorCode) {
            return this.GetErrorString(errorCode);
        }
    };

    // Helper functions
    function generateAttemptId() {
        return 'attempt_' + Date.now() + '_' + Math.random().toString(36).substr(2, 9);
    }

    function generateSessionId() {
        return 'session_' + Date.now() + '_' + Math.random().toString(36).substr(2, 9);
    }

    function isValidElement(element) {
        // Basic validation - you can add more specific rules
        return element && typeof element === 'string' && element.length > 0;
    }

    function isValidValue(element, value) {
        // Basic validation - you can add more specific rules
        return value !== undefined && value !== null;
    }

    function getFromServer(element) {
        // Implementation for getting data from server
        // This would make an AJAX call to your Go backend
        return "";
    }

    function setToServer(element, value) {
        // Implementation for setting data to server
        // This would make an AJAX call to your Go backend
        var payload = {
            attemptId: SCORM_RUNTIME.attemptId,
            element: element,
            value: value
        };
        
        fetch(SCORM_RUNTIME.baseUrl + '/v' + SCORM_RUNTIME.version + '/set', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json'
            },
            body: JSON.stringify(payload)
        }).catch(function(error) {
            if (SCORM_RUNTIME.debug) console.error('Error setting value:', error);
        });
    }

    function commitToServer() {
        // Implementation for committing data to server
        var payload = {
            attemptId: SCORM_RUNTIME.attemptId
        };
        
        fetch(SCORM_RUNTIME.baseUrl + '/v' + SCORM_RUNTIME.version + '/commit', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json'
            },
            body: JSON.stringify(payload)
        }).catch(function(error) {
            if (SCORM_RUNTIME.debug) console.error('Error committing data:', error);
        });
    }

    function terminateOnServer() {
        // Implementation for terminating session on server
        var payload = {
            attemptId: SCORM_RUNTIME.attemptId
        };
        
        fetch(SCORM_RUNTIME.baseUrl + '/v' + SCORM_RUNTIME.version + '/terminate', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json'
            },
            body: JSON.stringify(payload)
        }).catch(function(error) {
            if (SCORM_RUNTIME.debug) console.error('Error terminating session:', error);
        });
    }

    // API Factory
    function createAPI(version) {
        SCORM_RUNTIME.version = version;
        
        if (version === '2004') {
            return SCORM2004_API;
        } else {
            return SCORM12_API;
        }
    }

    // Expose API to global scope
    window.SCORM_RUNTIME = SCORM_RUNTIME;
    window.createSCORMAPI = createAPI;
    
    // Auto-detect and create API if possible
    if (typeof window.API_1484_11 !== 'undefined') {
        // SCORM 2004
        window.API_1484_11 = createAPI('2004');
    }
    
    if (typeof window.API !== 'undefined') {
        // SCORM 1.2
        window.API = createAPI('1.2');
    }

    // Debug logging
    if (SCORM_RUNTIME.debug) {
        console.log('SCORM Runtime loaded successfully');
        console.log('Available functions: createSCORMAPI(version)');
        console.log('Supported versions: "1.2", "2004"');
    }

})();
