import React, { useState } from 'react';
import { Button, Dropdown, Row, Col, Card, Modal, Form, FormSelect } from 'react-bootstrap';
import { useNavigate } from 'react-router-dom';

const mockApps = [
    { id: 1, name: 'App One' },
    { id: 2, name: 'App Two' },
    { id: 3, name: 'App Three' },
    { id: 4, name: 'App Four' },
    { id: 5, name: 'App Five' },
    { id: 6, name: 'App Six' },
];

const namespaces = ['default', 'production', 'staging'];

const ApplicationsPage: React.FC = () => {
    const navigate = useNavigate();
    const [apps, setApps] = useState(mockApps);
    const [showModal, setShowModal] = useState(false);
    const [namespace, setNamespace] = useState('default');
    const [app, setApp] = useState('');
    const [secretKey, setSecretKey] = useState('');
    const [loading, setLoading] = useState(false);
    const [selectedNamespace, setSelectedNamespace] = useState(namespaces[0]);

    const handleDelete = (id: number) => {
        setApps(apps.filter(app => app.id !== id));
    };

    const handleCreateApp = async (e: React.FormEvent) => {
        e.preventDefault();
        setLoading(true);
        try {
            const payload = { ns: namespace, key: secretKey, app };
            await new Promise(res => setTimeout(res, 1000));
            setApps([...apps, { id: Date.now(), name: app }]);
            setShowModal(false);
            setNamespace('default');
            setApp('');
            setSecretKey('');
        } finally {
            setLoading(false);
        }
    };

    return (
        <Card className="border-0 shadow-sm" style={{ minHeight: '80vh', borderRadius: 10 }}>
            <Card.Body>
                <Row className="align-items-center mb-4">
                    <Col><h4 className="mb-0">Your Applications</h4></Col>
                    <Col xs="auto">
                        <Button style={{ backgroundColor: '#1C608C', border: 'none' }} onClick={() => setShowModal(true)}>
                            Create App
                        </Button>
                    </Col>
                </Row>
                <Row className="mb-4">
                    <Col md={3} sm={6} xs={12}>
                        <FormSelect
                            value={selectedNamespace}
                            onChange={e => setSelectedNamespace(e.target.value)}
                            style={{ maxWidth: 250 }}
                        >
                            {namespaces.map(ns => (
                                <option key={ns} value={ns}>{ns}</option>
                            ))}
                        </FormSelect>
                    </Col>
                </Row>
                <Row xs={1} md={3} className="g-4">
                    {apps.map(app => (
                        <Col key={app.id}>
                            <Card
                                className="h-100"
                                style={{ cursor: 'pointer', background: '#fff', }}
                                onClick={() => navigate(`/dashboard/apps/${app.name}`)}
                            >
                                <Card.Body className="d-flex flex-column justify-content-between">
                                    <div className="d-flex justify-content-between align-items-start">
                                        <Card.Title>{app.name}</Card.Title>
                                        <Dropdown align="end" onClick={e => e.stopPropagation()}>
                                            <Dropdown.Toggle as="span" style={{ cursor: 'pointer', border: 'none', background: 'none', padding: 0 }}>
                                                <svg width="24" height="24" fill="#1C608C" viewBox="0 0 24 24">
                                                    <circle cx="5" cy="12" r="2" />
                                                    <circle cx="12" cy="12" r="2" />
                                                    <circle cx="19" cy="12" r="2" />
                                                </svg>
                                            </Dropdown.Toggle>
                                            <Dropdown.Menu >
                                                <Dropdown.Item onClick={() => handleDelete(app.id)}>Delete</Dropdown.Item>
                                            </Dropdown.Menu>
                                        </Dropdown>
                                    </div>
                                    <Card.Text className="text-muted mt-2">Namespace: {selectedNamespace}</Card.Text>
                                </Card.Body>
                            </Card>
                        </Col>
                    ))}
                </Row>
                <Modal show={showModal} onHide={() => setShowModal(false)} centered>
                    <Modal.Header closeButton>
                        <Modal.Title>Create App</Modal.Title>
                    </Modal.Header>
                    <Form onSubmit={handleCreateApp}>
                        <Modal.Body>
                            <Form.Group className="mb-3" controlId="formNamespace">
                                <Form.Label>Namespace</Form.Label>
                                <Form.Control
                                    type="text"
                                    value={namespace}
                                    onChange={e => setNamespace(e.target.value)}
                                    required
                                    placeholder="Enter namespace"
                                />
                            </Form.Group>
                            <Form.Group className="mb-3" controlId="formApp">
                                <Form.Label>App</Form.Label>
                                <Form.Control
                                    type="text"
                                    value={app}
                                    onChange={e => setApp(e.target.value)}
                                    required
                                    placeholder="Enter app name"
                                />
                            </Form.Group>
                            <Form.Group className="mb-3" controlId="formSecretKey">
                                <Form.Label>Secret Key</Form.Label>
                                <Form.Control
                                    type="text"
                                    value={secretKey}
                                    onChange={e => setSecretKey(e.target.value)}
                                    required
                                    placeholder="Enter secret key"
                                />
                            </Form.Group>
                        </Modal.Body>
                        <Modal.Footer>
                            <Button type="submit" style={{ backgroundColor: '#1C608C', border: 'none' }} disabled={loading}>
                                {loading ? 'Creating...' : 'Create'}
                            </Button>
                        </Modal.Footer>
                    </Form>
                </Modal>
            </Card.Body>
        </Card>
    );
};

export default ApplicationsPage; 