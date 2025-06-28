import React, { useState } from 'react';
import { Card, Accordion, Form, Button, Modal, ListGroup } from 'react-bootstrap';

const SettingsPage: React.FC = () => {
    const [showModal, setShowModal] = useState(false);
    const [showChangePasswordModal, setShowChangePasswordModal] = useState(false);
    const [selectedSetting, setSelectedSetting] = useState('');
    const [pwd, setPwd] = useState('');
    const [radioValue, setRadioValue] = useState('option1');
    const [switchOn, setSwitchOn] = useState(false);
    const [bucketHost, setBucketHost] = useState('');
    const [s3AccessKey, setS3AccessKey] = useState('');
    const [s3SecretKey, setS3SecretKey] = useState('');
    const [checkingS3, setCheckingS3] = useState(false);
    const [savingS3, setSavingS3] = useState(false);
    const [loading, setLoading] = useState(false);


    const handleOpenModal = (setting: string) => {
        setSelectedSetting(setting);
        setShowModal(true);
    };

    const handleCloseModal = () => {
        setShowModal(false);
        setSelectedSetting('');
    };

    const handleCheckS3 = () => {
        setCheckingS3(true);
        // Implementation of handleCheckS3
    };

    const handleSaveS3 = () => {
        setSavingS3(true);
        // Implementation of handleSaveS3
    };

    const handleChangePassword = () => {
        setLoading(true);
        // Implementation of handleChangePassword
        setLoading(false);
        setShowChangePasswordModal(true);
    };

    return (
        <Card className="border-0 shadow-sm" style={{ minHeight: '80vh', borderRadius: 10 }}>
            <Card.Body>
                <h4 className="mb-4">Settings</h4>
                <Accordion defaultActiveKey="0">
                    <Accordion.Item eventKey="0">
                        <Accordion.Header>S3 Backup & Fallback</Accordion.Header>
                        <Accordion.Body>
                            <Form>
                                {/* <Form.Check
                                    type="radio"
                                    label="Light"
                                    name="themeRadios"
                                    id="theme-light"
                                    value="option1"
                                    checked={radioValue === 'option1'}
                                    onChange={() => setRadioValue('option1')}
                                />
                                <Form.Check
                                    type="radio"
                                    label="Dark"
                                    name="themeRadios"
                                    id="theme-dark"
                                    value="option2"
                                    checked={radioValue === 'option2'}
                                    onChange={() => setRadioValue('option2')}
                                />
                                <hr className="my-4" /> */}
                                <Form.Group className="mb-3" controlId="formS3AccessKey">
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
                                    <Button style={{ border: '1px solid#1C608C', backgroundColor: '#FEF8EE', color: 'black' }} onClick={handleCheckS3} disabled={checkingS3}>
                                        {checkingS3 ? 'Checking...' : 'Check'}
                                    </Button>
                                    <Button style={{ backgroundColor: '#1C608C', border: 'none' }} onClick={handleSaveS3} disabled={savingS3}>
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
                                {/* <ListGroup.Item action onClick={() => handleOpenModal('Advanced Setting 2')}>
                                    Advanced Setting 2
                                </ListGroup.Item> */}
                            </ListGroup>
                        </Accordion.Body>
                    </Accordion.Item>
                </Accordion>
                <Modal show={showModal} onHide={handleCloseModal} centered>
                    <Modal.Header closeButton>
                        <Modal.Title>{selectedSetting}</Modal.Title>
                    </Modal.Header>
                    <Modal.Body>
                        <p>Settings for <b>{selectedSetting}</b> go here.</p>
                        <Button variant="primary" onClick={handleCloseModal}>Close</Button>
                    </Modal.Body>
                </Modal>

                <Modal show={showChangePasswordModal} onHide={() => setShowChangePasswordModal(false)} centered>
                    <Modal.Header closeButton>
                        <Modal.Title>Change Default Password</Modal.Title>
                    </Modal.Header>
                    <Form onSubmit={handleChangePassword}>
                        <Modal.Body>
                            <Form.Group className="mb-3" controlId="formNamespace">
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

export default SettingsPage; 