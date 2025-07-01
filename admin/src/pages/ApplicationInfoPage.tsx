import React, { useEffect, useState } from 'react';
import { Card, Tabs, Tab, Form, Button, Accordion, ListGroup, Modal, ButtonGroup, Badge } from 'react-bootstrap';
import Editor from '@monaco-editor/react';
import { useAuth } from '../contexts/AuthContext';
import { useLocation } from 'react-router-dom';

interface Secret {
    name: string;
    value: string;
}

interface Deployment {
    description: string;
    version: string;
    deployed: boolean;
    diff: string;
}

const ApplicationInfoPage: React.FC = () => {
    const [jsonConfig, setJsonConfig] = useState('{\n  "example": "configuration",\n  "settings": {\n    "enabled": true,\n    "timeout": 5000\n  }\n}');
    const [jsonSecret, setJsonSecret] = useState('{\n  "example": "secret",\n  "settings": {\n    "enabled": true,\n    "timeout": 5000\n  }\n}');
    const [showRulesModal, setShowRulesModal] = useState(false);
    const [rulesContent, setRulesContent] = useState('// Add your JavaScript rules here\n\nfunction validateConfig(config) {\n  // Example rule\n  if (!config.enabled) {\n    return false;\n  }\n  return true;\n}');
    // const [switchOn, setSwitchOn] = useState(false);
    const [bucketHost, setBucketHost] = useState('');
    // const [s3AccessKey, setS3AccessKey] = useState('');
    // const [s3SecretKey, setS3SecretKey] = useState('');
    const [checkingS3, setCheckingS3] = useState(false);
    const [savingS3, setSavingS3] = useState(false);
    const [pwd, setPwd] = useState('');
    const [loading, setLoading] = useState(false);
    const [showChangePasswordModal, setShowChangePasswordModal] = useState(false);
    const [savingRules, setSavingRules] = useState(false);
    const [deploymentDesc, setDeploymentDesc] = useState('');
    const [accessKey, setAccessKey] = useState('');
    const [secretKey, setSecretKey] = useState('');
    const [secrets, setSecrets] = useState<Secret[]>([]);
    const [deletedSecrets, setDeletedSecrets] = useState<Secret[]>([]);
    const [deployments, setDeployments] = useState<Deployment[]>([]);

    const { hasRole, hasPermission } = useAuth();
    const { ns, app } = useLocation().state;

    useEffect(() => {
        fetchConfig();
    }, []);

    const fetchConfig = async () => {
        const payload = { ns, app: app.replace(ns + '-', '') };
        const response = await fetch(`http://localhost:7766/api/v1/state?${new URLSearchParams(payload).toString()}`);
        const data = await response.json();
        setJsonConfig(JSON.stringify(data.configs, null, 2));
        setAccessKey(data.access_key);
        setSecretKey(data.secret_key);

        setSecrets(data.secrets);
        setDeployments(data.deployments);
    }

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

    const handleCreateDeployment = async () => {
        // we need to first validate with the schema
    };

    const handleSaveSecrets = async () => {
        if (secrets.length === 0 && deletedSecrets.length === 0) {
            return;
        }

        const payload = {
            secrets: secrets,
            deleted_secrets: deletedSecrets
        }

        await fetch(`http://localhost:7766/api/v1/admin.secrets?${new URLSearchParams({ ns, app: app.replace(ns + '-', '') }).toString()}`, {
            method: 'PUT',
            headers: {
                'Content-Type': 'application/json',
            },
            body: JSON.stringify(payload)
        })
    }

    const canEditPodUrl = hasRole('admin') || (hasRole('user') && hasPermission('update_pod_url'));

    return (
        <Card className="border-0 shadow-sm" style={{ minHeight: '80vh', borderRadius: 10 }}>
            <Card.Body>
                <h4 className="mb-4">{app}</h4>

                <Tabs defaultActiveKey="info" className="mb-3">
                    <Tab eventKey="info" title="Info">
                        <Accordion defaultActiveKey="0">
                            <Accordion.Item eventKey="0">
                                <Accordion.Header>Access Key</Accordion.Header>
                                <Accordion.Body>
                                    <div style={{ backgroundColor: '#FEFBF0', padding: 10, borderRadius: 10 }} className="d-flex justify-content-between align-items-center">
                                        {accessKey}

                                        <Button style={{ backgroundColor: '#1C608C', border: 'none' }}>
                                            Copy
                                        </Button>
                                    </div>
                                </Accordion.Body>
                            </Accordion.Item>

                            <Accordion.Item eventKey="1">
                                <Accordion.Header>Secret Key</Accordion.Header>
                                <Accordion.Body>
                                    <div style={{ backgroundColor: '#FEFBF0', padding: 10, borderRadius: 10 }} className="d-flex justify-content-between align-items-center">
                                        {secretKey}

                                        <Button style={{ backgroundColor: '#1C608C', border: 'none' }}>
                                            Copy
                                        </Button>
                                    </div>
                                </Accordion.Body>
                            </Accordion.Item>

                            <Accordion.Item eventKey="2">
                                <Accordion.Header>Deployment Healthcheck</Accordion.Header>
                                <Accordion.Body>
                                    <div>
                                        Status: Cool
                                    </div>
                                </Accordion.Body>
                            </Accordion.Item>

                            {canEditPodUrl ? <Accordion.Item eventKey="2">
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
                                                style={{ border: '1px solid #1C608C', backgroundColor: '#FEFBF0', color: 'black' }}
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

                    <Tab eventKey="configure" title="Configs">
                        <Card className="border-0 shadow-sm">
                            <Card.Body>
                                <div className="d-flex justify-content-between align-items-start mb-3">
                                    <h5 className="mb-0">Edit Configuration</h5>
                                    {hasPermission('edit_schema') ? <Button
                                        style={{
                                            borderRadius: '8px',
                                            display: 'flex',
                                            alignItems: 'center',
                                            justifyContent: 'center',
                                            border: '1px solid #1C608C',
                                            backgroundColor: '#FEFBF0',
                                            color: '#1C608C'
                                        }}
                                        onClick={() => setShowRulesModal(true)}
                                    >
                                        🔒 Edit Schema
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
                                    <Form onSubmit={handleCreateDeployment}>
                                        <Form.Group className="mb-3 mt-2" controlId="formsDeploymentDesc">
                                            {/* <Form.Label>Deployment Description</Form.Label> */}
                                            <Form.Control
                                                minLength={20}
                                                required
                                                type="text"
                                                placeholder="Enter deployment description"
                                                value={deploymentDesc}
                                                onChange={e => setDeploymentDesc(e.target.value)}
                                            />
                                        </Form.Group>
                                    </Form>
                                    <div className="d-flex justify-content-end">
                                        <Button type='submit' className='mt-2' style={{ backgroundColor: '#1C608C', border: 'none' }} >
                                            Create Deployment
                                        </Button>
                                    </div>
                                </div> : <></>}
                            </Card.Body>
                        </Card>
                    </Tab>

                    {/* secrets tab */}
                    <Tab eventKey="secrets" title="Secrets">
                        <Card className="border-0 shadow-sm">
                            <Card.Body>
                                <h5 className="mb-4">Edit Secrets</h5>

                                {secrets.length === 0 ? <ListGroup.Item className="text-center text-muted">
                                    No secrets here!
                                </ListGroup.Item> : <></>}

                                {secrets.map((secret, index) => (
                                    <Form>
                                        <Form.Group className="d-flex col mb-3 gap-2" controlId="formsDeploymentDesc">
                                            {/* <Form.Label>Deployment Description</Form.Label> */}
                                            <Form.Control
                                                minLength={20}
                                                type="text"
                                                placeholder="Enter secret"
                                                value={secret.name}
                                                onChange={e => {
                                                    setSecrets(secrets.map((s, i) => i === index ? { ...s, name: e.target.value } : s));
                                                }}
                                            />

                                            <Form.Control
                                                minLength={20}
                                                type="text"
                                                placeholder="Enter value"
                                                value={secret.value}
                                                onChange={e => {
                                                    setSecrets(secrets.map((s, i) => i === index ? { ...s, value: e.target.value } : s));
                                                }}
                                            />

                                            {hasPermission('edit_schema') ? <Button
                                                style={{
                                                    borderRadius: '8px',
                                                    display: 'flex',
                                                    alignItems: 'center',
                                                    justifyContent: 'center',
                                                    border: '1px solid #1C608C',
                                                    backgroundColor: '#FEFBF0',
                                                    color: '#1C608C'
                                                }}
                                                onClick={() => setShowRulesModal(true)}
                                            >
                                                🔑
                                            </Button> : <></>}

                                            {hasPermission('edit_schema') ? <Button
                                                variant='danger'
                                                style={{
                                                    borderRadius: '8px',
                                                    display: 'flex',
                                                    alignItems: 'center',
                                                    justifyContent: 'center',
                                                }}
                                                onClick={() => {
                                                    // add to deleted secrets if the name and value are not empty
                                                    if (secrets[index].name !== '' || secrets[index].value !== '') {
                                                        setDeletedSecrets([...deletedSecrets, secrets[index]]);
                                                    }

                                                    // remove from main secrets list
                                                    setSecrets(secrets.filter((_, i) => i !== index));
                                                }}
                                            >
                                                Delete
                                            </Button> : <></>}
                                        </Form.Group>
                                    </Form>

                                ))}

                                <div className="mt-3 d-flex justify-content-end gap-2">
                                    <Button onClick={() => {
                                        setSecrets([...secrets, { name: '', value: '' }]);
                                    }} style={{ backgroundColor: '#1C608C', border: 'none' }}>+ Add</Button>

                                    <Button onClick={handleSaveSecrets} style={{ backgroundColor: '#1C608C', border: 'none' }}>Save</Button>
                                </div>


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

                    <Tab eventKey="deployments" title="Deployments">
                        <ListGroup>
                            {deployments.map(d => <ListGroup.Item className="d-flex justify-content-between align-items-center">
                                <div className="flex-grow-1">
                                    <div className="d-flex justify-content-between align-items-center">
                                        <div>
                                            <h6 className="mb-0">ver. {d.version}</h6>
                                            <div className="text-muted small">
                                                {d.description}
                                            </div>
                                        </div>
                                    </div>
                                </div>

                                {hasRole('user') && hasPermission('deploy_configs') && !d.deployed ? <Button
                                    style={{ backgroundColor: '#1C608C', border: 'none' }}
                                    size="sm"
                                // onClick={() => handleRevokeUser(user.id)}
                                // style={{ marginLeft: '1rem' }}
                                >
                                    Deploy
                                </Button> : <Badge bg="success">Deployed</Badge>}

                            </ListGroup.Item>)}
                        </ListGroup>
                    </Tab>
                </Tabs>

                {/* Rules Modal */}
                <Modal style={{ fontFamily: 'Quicksand' }} show={showRulesModal} onHide={() => setShowRulesModal(false)} centered size="lg">
                    <Modal.Header closeButton>
                        <Modal.Title>Schema</Modal.Title>
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