import React, { useState } from 'react';
import { Card, Tabs, Tab, Form, Button, Accordion, ListGroup, Modal } from 'react-bootstrap';
import Editor from '@monaco-editor/react';
import { useAuth } from '../contexts/AuthContext';

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
    const [deploymentDesc, setDeploymentDesc] = useState('');

    const { hasRole, hasPermission } = useAuth();

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

    const canEditPodUrl = hasRole('admin') || (hasRole('user') && hasPermission('update_pod_url'));

    return (
        <Card className="border-0 shadow-sm" style={{ minHeight: '80vh', borderRadius: 10 }}>
            <Card.Body>
                <h4 className="mb-4">Application State</h4>

                <Tabs defaultActiveKey="configure" className="mb-3">
                    <Tab eventKey="configure" title="Configs">
                        <Card className="border-0 shadow-sm">
                            <Card.Body>
                                <div className="d-flex justify-content-between align-items-start mb-3">
                                    <h5 className="mb-0">Edit Configuration</h5>
                                    {hasPermission('edit_config_rule') ? <Button
                                        style={{
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
                                        🔒 Edit Rules
                                    </Button> : <></>}
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

                                {hasRole('user') && hasPermission('create_configs') ? <div>
                                    <Form>
                                        <Form.Group className="mb-3 mt-2" controlId="formsDeploymentDesc">
                                            {/* <Form.Label>Deployment Description</Form.Label> */}
                                            <Form.Control
                                                minLength={20}
                                                type="text"
                                                placeholder="Enter deployment description"
                                                value={deploymentDesc}
                                                onChange={e => setDeploymentDesc(e.target.value)}
                                            />
                                        </Form.Group>
                                    </Form>
                                    <div className="d-flex justify-content-end">
                                        <Button style={{ backgroundColor: '#1C608C', border: 'none', marginTop: 10 }} >
                                            Create Deployment
                                        </Button>
                                    </div>
                                </div> : <></>}
                            </Card.Body>
                        </Card>
                    </Tab>
                    <Tab eventKey="secrets" title="Secrets">
                        <Card className="border-0 shadow-sm">
                            <Card.Body>
                                <div className="d-flex justify-content-between align-items-start mb-3">
                                    <h5 className="mb-0">Edit Secrets</h5>
                                    {hasPermission('edit_secret_policy') ? <Button
                                        style={{
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
                                        🔑 Edit Policy
                                    </Button> : <></>}
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

                                {hasRole('user') && hasPermission('create_secrets') ? <div>
                                    <Form>
                                        <Form.Group className="mb-3 mt-2" controlId="formsDeploymentDesc">
                                            {/* <Form.Label>Deployment Description</Form.Label> */}
                                            <Form.Control
                                                minLength={20}
                                                type="text"
                                                placeholder="Enter deployment description"
                                                value={deploymentDesc}
                                                onChange={e => setDeploymentDesc(e.target.value)}
                                            />
                                        </Form.Group>
                                    </Form>
                                    <div className="d-flex justify-content-end">
                                        <Button style={{ backgroundColor: '#1C608C', border: 'none', marginTop: 10 }} >
                                            Create Deployment
                                        </Button>
                                    </div>
                                </div> : <></>}
                            </Card.Body>


                        </Card>
                    </Tab>

                    <Tab eventKey="settings" title="Status">
                        <Accordion defaultActiveKey="0">
                            <Accordion.Item eventKey="0">
                                <Accordion.Header>Deployment Healthcheck</Accordion.Header>
                                <Accordion.Body>
                                    <div>
                                        Status: Cool
                                    </div>
                                </Accordion.Body>
                            </Accordion.Item>

                            {canEditPodUrl ? <Accordion.Item eventKey="1">
                                <Accordion.Header>Pod Details</Accordion.Header>
                                <Accordion.Body>
                                    <Form>
                                        <Form.Group className="mb-3" controlId="formS3Host">
                                            <Form.Label>Pod URL</Form.Label>
                                            <Form.Control
                                                disabled={!canEditPodUrl}
                                                type="text"
                                                placeholder="Enter your application host URL"
                                                value={bucketHost}
                                                onChange={e => setBucketHost(e.target.value)}
                                            />
                                        </Form.Group>
                                        <div className="d-flex gap-2">
                                            <Button
                                                style={{ border: '1px solid#1C608C', backgroundColor: '#FEF8EE', color: 'black' }}
                                                onClick={handleCheckS3}
                                                disabled={checkingS3}
                                            >
                                                {checkingS3 ? 'Verifying...' : 'Verify'}
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
                            </Accordion.Item> : <></>}
                        </Accordion>
                    </Tab>

                    <Tab eventKey="deployments" title="Deployments">
                        <Accordion defaultActiveKey="0">
                            <Accordion.Item eventKey="0">
                                <Accordion.Header>Configs</Accordion.Header>
                                <Accordion.Body>
                                    <ListGroup>
                                        <ListGroup.Item className="d-flex justify-content-between align-items-center">
                                            <div className="flex-grow-1">
                                                <div className="d-flex justify-content-between align-items-center">
                                                    <div>
                                                        <h6 className="mb-0">Deployment 1</h6>
                                                        <div className="text-muted small">
                                                            this is a description of the deployment
                                                        </div>
                                                    </div>
                                                </div>
                                            </div>

                                            {(hasRole('admin') || hasRole('user')) && hasPermission('deploy_configs') ? <Button
                                                style={{ backgroundColor: '#1C608C', border: 'none' }}
                                                size="sm"
                                            // onClick={() => handleRevokeUser(user.id)}
                                            // style={{ marginLeft: '1rem' }}
                                            >
                                                Deploy
                                            </Button> : <></>}

                                        </ListGroup.Item>
                                    </ListGroup>
                                </Accordion.Body>
                            </Accordion.Item>

                            <Accordion.Item eventKey="1">
                                <Accordion.Header>Secrets</Accordion.Header>
                                <Accordion.Body>
                                    <ListGroup>
                                        <ListGroup.Item className="d-flex justify-content-between align-items-center">
                                            <div className="flex-grow-1">
                                                <div className="d-flex justify-content-between align-items-center">
                                                    <div>
                                                        <h6 className="mb-0">Deployment 1</h6>
                                                        <div className="text-muted small">
                                                            this is a description of the deployment
                                                        </div>
                                                    </div>
                                                </div>
                                            </div>

                                            {(hasRole('admin') || hasRole('user')) && hasPermission('deploy_secrets') ? <Button
                                                style={{ backgroundColor: '#1C608C', border: 'none' }}
                                                size="sm"
                                            // onClick={() => handleRevokeUser(user.id)}
                                            // style={{ marginLeft: '1rem' }}
                                            >
                                                Deploy
                                            </Button> : <></>}

                                        </ListGroup.Item>
                                    </ListGroup>
                                </Accordion.Body>
                            </Accordion.Item>
                        </Accordion>
                    </Tab>
                </Tabs>

                {/* Rules Modal */}
                <Modal style={{ fontFamily: 'Quicksand' }} show={showRulesModal} onHide={() => setShowRulesModal(false)} centered size="lg">
                    <Modal.Header closeButton>
                        <Modal.Title>Rules Configuration</Modal.Title>
                    </Modal.Header>
                    <Modal.Body>
                        <Form.Group controlId="rulesEditor">
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
                            style={{ backgroundColor: '#1C608C', border: 'none' }}
                            onClick={handleSaveRules}
                            disabled={savingRules}
                        >
                            {savingRules ? 'Saving...' : 'Save Rules'}
                        </Button>
                    </Modal.Footer>
                </Modal>

                {/* Change Password Modal */}
                <Modal style={{ fontFamily: 'Quicksand' }} show={showChangePasswordModal} onHide={() => setShowChangePasswordModal(false)} centered>
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