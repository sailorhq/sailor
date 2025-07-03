import React, { useEffect, useState } from 'react';
import { Button, Row, Col, Card, Modal, Form, FormSelect } from 'react-bootstrap';
import { useNavigate } from 'react-router-dom';
import { useAuth } from '../contexts/AuthContext';
import { useDispatch } from 'react-redux';
import { setApps as setAppsAction } from '../appsSlice';
import { useSelector } from 'react-redux';
import type { RootState } from '../store';



const ApplicationsPage: React.FC = () => {
    const navigate = useNavigate();
    const [showModal, setShowModal] = useState(false);
    const [namespaceToCreate, setNamespaceToCreate] = useState('');
    const [namespaces, setNamespaces] = useState(['default']);
    const [app, setApp] = useState('');
    const [secretKey, setSecretKey] = useState('');
    const [loading, setLoading] = useState(false);
    const [selectedNamespace, setSelectedNamespace] = useState('default');

    const { user, hasRole } = useAuth();

    const dispatch = useDispatch();

    useEffect(() => {
        fetchApps();
    }, [])

    const fetchApps = async () => {
        const res = await fetch('http://localhost:7766/api/v1/admin.list.apps', {
            headers: {
                'x-token': user?.token || '',
                'x-username': user?.username || '',
            },
        });
        if (!res.ok) {
        }
        const data = await res.json() as string[];
        const nsListSet = new Set([...namespaces, ...data.map(project => project.split('-')[0])]);
        setNamespaces(Array.from(nsListSet));

        dispatch(setAppsAction(data));
    }

    const handleCreateApp = async (e: React.FormEvent) => {
        e.preventDefault();
        setLoading(false);
        const payload = { ns: namespaceToCreate, key: secretKey, app };
        const res = await fetch(`http://localhost:7766/api/v1/create?${new URLSearchParams(payload).toString()}`, {
            method: 'PUT',
            headers: {
                'x-username': user?.username || '',
            },
        });

        if (res.ok) {
            setShowModal(false);
            setNamespaceToCreate('');
            setApp('');
            setSecretKey('');


            fetchApps();
        }
    };

    return (
        <Card className="border-0 shadow-sm" style={{ minHeight: '80vh', borderRadius: 10 }}>
            <Card.Body>
                {hasRole('admin') ? <Row className="align-items-center mb-4">
                    <Col><h4 className="mb-0">Your Applications</h4></Col>
                    <Col xs="auto">
                        <Button style={{ backgroundColor: '#1C608C', border: 'none' }} onClick={() => setShowModal(true)}>
                            Create App
                        </Button>
                    </Col>
                </Row> : <></>}
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
                    {useSelector((state: RootState) => state.apps.values.filter(app => app.startsWith(selectedNamespace))).map((app: string) => (
                        <Col key={app}>
                            <Card
                                className="h-100"
                                style={{ cursor: hasRole('admin') ? 'default' : 'pointer', background: '#fff', }}
                                onClick={hasRole('admin') ? undefined : () => navigate(`/dashboard/apps/${app}`, { state: { ns: selectedNamespace, app: app } })}
                            >
                                <Card.Body className="d-flex flex-column justify-content-between">
                                    <div className="d-flex justify-content-between align-items-start">
                                        <Card.Title>{app}</Card.Title>
                                    </div>
                                    {/* <Card.Text className="text-muted small">version: {selectedNamespace}</Card.Text> */}
                                </Card.Body>
                            </Card>
                        </Col>
                    ))}
                </Row>
                <Modal style={{ fontFamily: 'Quicksand' }} show={showModal} onHide={() => setShowModal(false)} centered>
                    <Modal.Header closeButton>
                        <Modal.Title>Create App</Modal.Title>
                    </Modal.Header>
                    <Form onSubmit={handleCreateApp}>
                        <Modal.Body>
                            <Form.Group className="mb-3" controlId="formNamespace">
                                <Form.Label>Namespace</Form.Label>
                                <Form.Control
                                    type="text"
                                    value={namespaceToCreate}
                                    onChange={e => setNamespaceToCreate(e.target.value)}
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