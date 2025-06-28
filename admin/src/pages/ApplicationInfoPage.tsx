import React, { useState } from 'react';
import { Card, Tabs, Tab, Form, Button, Accordion, ListGroup, Modal } from 'react-bootstrap';
import Editor from '@monaco-editor/react';

const ApplicationInfoPage: React.FC = () => {
    const [jsonConfig, setJsonConfig] = useState('{\n  "example": "configuration",\n  "settings": {\n    "enabled": true,\n    "timeout": 5000\n  }\n}');
    const [showRulesModal, setShowRulesModal] = useState(false);
    const [rulesContent, setRulesContent] = useState('// Add your JavaScript rules here\n\nfunction validateConfig(config) {\n  // Example rule\n  if (!config.enabled) {\n    return false;\n  }\n  return true;\n}');
    const [switchOn, setSwitchOn] = useState(false);
    const [bucketHost, setBucketHost] = useState('');
    const [s3AccessKey, setS3AccessKey] = useState('');
    const [s3SecretKey, setS3SecretKey] = useState('');
    const [checkingS3, setCheckingS3] = useState(false);
    const [savingS3, setSavingS3] = useState(false);
    const [pwd, setPwd] = useState('');
    const [loading, setLoading] = useState(false);
    const [showChangePasswordModal, setShowChangePasswordModal] = useState(false);
    const [savingRules, setSavingRules] = useState(false);

    const handleCheckS3 = () => {
        setCheckingS3(true);
        // Implementation of handleCheckS3
        setTimeout(() => setCheckingS3(false), 2000);
    };

    const handleSaveS3 = () => {
        setSavingS3(true);
        // Implementation of handleSaveS3
        setTimeout(() => setSavingS3(false), 2000);
    };

    const handleChangePassword = () => {
        setLoading(true);
        // Implementation of handleChangePassword
        setTimeout(() => {
            setLoading(false);
            setShowChangePasswordModal(true);
        }, 1000);
    };

    const handleSaveRules = async () => {
        setSavingRules(true);
        try {
            // API call to save rules
            const response = await fetch('/api/rules', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                },
                body: JSON.stringify({
                    rules: rulesContent,
                    language: 'javascript'
                })
            });

            if (response.ok) {
                // Success handling
                console.log('Rules saved successfully');
                setShowRulesModal(false);
            } else {
                // Error handling
                console.error('Failed to save rules');
            }
        } catch (error) {
            console.error('Error saving rules:', error);
        } finally {
            setSavingRules(false);
        }
    };

    return (
        <Card className="border-0 shadow-sm" style={{ background: '#FEF8EE', minHeight: '80vh' }}>
            <Card.Body style={{ background: '#FEF8EE' }}>
                <h4 className="mb-4">Application State</h4>

                <Tabs defaultActiveKey="configure" className="mb-3">
                    <Tab eventKey="configure" title="Configs">
                        <Card className="border-0 shadow-sm">
                            <Card.Body>
                                <div className="d-flex justify-content-between align-items-start mb-3">
                                    <h5 className="mb-0">Configuration</h5>
                                    <Button
                                        variant="outline-primary"
                                        size="sm"
                                        style={{
                                            width: '40px',
                                            height: '40px',
                                            borderRadius: '8px',
                                            display: 'flex',
                                            alignItems: 'center',
                                            justifyContent: 'center',
                                            border: '1px solid #1C608C',
                                            backgroundColor: '#FEF8EE',
                                            color: '#1C608C'
                                        }}
                                        onClick={() => setShowRulesModal(true)}
                                    >
                                        ⚙️
                                    </Button>
                                </div>
                                <Form.Group controlId="jsonConfig">
                                    {/* <Form.Label>JSON Configuration</Form.Label> */}
                                    <div style={{ height: '400px', border: '1px solid #ced4da', borderRadius: '4px' }}>
                                        <Editor
                                            height="100%"
                                            language="json"
                                            theme="vs"
                                            value={jsonConfig}
                                            onChange={(value) => setJsonConfig(value || '')}
                                            options={{
                                                minimap: { enabled: false },
                                                scrollBeyondLastLine: false,
                                                fontSize: 14,
                                                fontFamily: 'monospace',
                                                automaticLayout: true,
                                                wordWrap: 'on'
                                            }}
                                        />
                                    </div>
                                </Form.Group>
                            </Card.Body>
                        </Card>
                    </Tab>
                    <Tab eventKey="secrets" title="Secrets">
                        <Card className="border-0 shadow-sm">
                            <Card.Body>
                                <div className="d-flex justify-content-between align-items-start mb-3">
                                    <h5 className="mb-0">Edit Secrets</h5>
                                    <Button
                                        variant="outline-primary"
                                        size="sm"
                                        style={{
                                            width: '40px',
                                            height: '40px',
                                            borderRadius: '8px',
                                            display: 'flex',
                                            alignItems: 'center',
                                            justifyContent: 'center',
                                            border: '1px solid #1C608C',
                                            backgroundColor: '#FEF8EE',
                                            color: '#1C608C'
                                        }}
                                        onClick={() => setShowRulesModal(true)}
                                    >
                                        ⚙️
                                    </Button>
                                </div>
                                <Form.Group controlId="jsonConfig">
                                    {/* <Form.Label>JSON Configuration</Form.Label> */}
                                    <div style={{ height: '400px', border: '1px solid #ced4da', borderRadius: '4px' }}>
                                        <Editor
                                            height="100%"
                                            language="json"
                                            theme="vs"
                                            value={jsonConfig}
                                            onChange={(value) => setJsonConfig(value || '')}
                                            options={{
                                                minimap: { enabled: false },
                                                scrollBeyondLastLine: false,
                                                fontSize: 14,
                                                fontFamily: 'monospace',
                                                automaticLayout: true,
                                                wordWrap: 'on'
                                            }}
                                        />
                                    </div>
                                </Form.Group>
                            </Card.Body>
                        </Card>
                    </Tab>

                    <Tab eventKey="settings" title="App Settings">
                        <Accordion defaultActiveKey="0">
                            <Accordion.Item eventKey="0">
                                <Accordion.Header>S3 Backup & Fallback</Accordion.Header>
                                <Accordion.Body>
                                    <Form>
                                        <Form.Group className="mb-3" controlId="formS3Host">
                                            <Form.Label>S3 Host</Form.Label>
                                            <Form.Control
                                                type="text"
                                                placeholder="Enter S3 Host"
                                                value={bucketHost}
                                                onChange={e => setBucketHost(e.target.value)}
                                            />
                                        </Form.Group>
                                        <Form.Group className="mb-3" controlId="formS3AccessKey">
                                            <Form.Label>S3 Access Key</Form.Label>
                                            <Form.Control
                                                type="text"
                                                placeholder="Enter S3 Access Key"
                                                value={s3AccessKey}
                                                onChange={e => setS3AccessKey(e.target.value)}
                                            />
                                        </Form.Group>
                                        <Form.Group className="mb-3" controlId="formS3SecretKey">
                                            <Form.Label>S3 Secret Key</Form.Label>
                                            <Form.Control
                                                type="password"
                                                placeholder="Enter S3 Secret Key"
                                                value={s3SecretKey}
                                                onChange={e => setS3SecretKey(e.target.value)}
                                            />
                                        </Form.Group>
                                        <div className="d-flex gap-2">
                                            <Button
                                                style={{ border: '1px solid#1C608C', backgroundColor: '#FEF8EE', color: 'black' }}
                                                onClick={handleCheckS3}
                                                disabled={checkingS3}
                                            >
                                                {checkingS3 ? 'Checking...' : 'Check'}
                                            </Button>
                                            <Button
                                                style={{ backgroundColor: '#1C608C', border: 'none' }}
                                                onClick={handleSaveS3}
                                                disabled={savingS3}
                                            >
                                                {savingS3 ? 'Saving...' : 'Save'}
                                            </Button>
                                        </div>
                                    </Form>
                                </Accordion.Body>
                            </Accordion.Item>
                            <Accordion.Item eventKey="1">
                                <Accordion.Header>Logging</Accordion.Header>
                                <Accordion.Body>
                                    <Form>
                                        <Form.Check
                                            type="switch"
                                            id="custom-switch"
                                            label="Enable Deployment Logging"
                                            checked={switchOn}
                                            onChange={() => setSwitchOn(!switchOn)}
                                        />
                                    </Form>
                                </Accordion.Body>
                            </Accordion.Item>
                            <Accordion.Item eventKey="2">
                                <Accordion.Header>Panel</Accordion.Header>
                                <Accordion.Body>
                                    <ListGroup variant="flush">
                                        <ListGroup.Item action onClick={() => handleChangePassword()}>
                                            Change Default Password
                                        </ListGroup.Item>
                                    </ListGroup>
                                </Accordion.Body>
                            </Accordion.Item>
                        </Accordion>
                    </Tab>
                </Tabs>

                {/* Rules Modal */}
                <Modal show={showRulesModal} onHide={() => setShowRulesModal(false)} centered size="lg">
                    <Modal.Header closeButton>
                        <Modal.Title>Rules Configuration</Modal.Title>
                    </Modal.Header>
                    <Modal.Body>
                        <Form.Group controlId="rulesEditor">
                            <Form.Label>Config Rules</Form.Label>
                            <div style={{ height: '400px', border: '1px solid #ced4da', borderRadius: '4px' }}>
                                <Editor
                                    height="100%"
                                    language="javascript"
                                    theme="vs"
                                    value={rulesContent}
                                    onChange={(value) => setRulesContent(value || '')}
                                    options={{
                                        minimap: { enabled: false },
                                        scrollBeyondLastLine: false,
                                        fontSize: 14,
                                        fontFamily: 'monospace',
                                        automaticLayout: true,
                                        wordWrap: 'on',
                                        lineNumbers: 'on',
                                        roundedSelection: false,
                                        readOnly: false,
                                        cursorStyle: 'line',
                                        contextmenu: true,
                                        mouseWheelZoom: true,
                                        suggestOnTriggerCharacters: true,
                                        acceptSuggestionOnEnter: 'on'
                                    }}
                                />
                            </div>
                        </Form.Group>
                    </Modal.Body>
                    <Modal.Footer>
                        <Button
                            variant="secondary"
                            onClick={() => setShowRulesModal(false)}
                        >
                            Close
                        </Button>
                        <Button
                            style={{ backgroundColor: '#1C608C', border: 'none' }}
                            onClick={handleSaveRules}
                            disabled={savingRules}
                        >
                            {savingRules ? 'Saving...' : 'Save Rules'}
                        </Button>
                    </Modal.Footer>
                </Modal>

                {/* Change Password Modal */}
                <Modal show={showChangePasswordModal} onHide={() => setShowChangePasswordModal(false)} centered>
                    <Modal.Header closeButton>
                        <Modal.Title>Change Default Password</Modal.Title>
                    </Modal.Header>
                    <Form onSubmit={handleChangePassword}>
                        <Modal.Body>
                            <Form.Group className="mb-3" controlId="formPassword">
                                <Form.Label>Password</Form.Label>
                                <Form.Control
                                    type="text"
                                    value={pwd}
                                    onChange={e => setPwd(e.target.value)}
                                    required
                                    placeholder="Enter password"
                                />
                            </Form.Group>
                        </Modal.Body>
                        <Modal.Footer>
                            <Button
                                type="submit"
                                style={{ backgroundColor: '#1C608C', border: 'none' }}
                                disabled={loading}
                            >
                                {loading ? 'Creating...' : 'Create'}
                            </Button>
                        </Modal.Footer>
                    </Form>
                </Modal>
            </Card.Body>
        </Card>
    );
};

export default ApplicationInfoPage;