import React, { useEffect, useState } from 'react';
import { Toast as BootstrapToast, ToastContainer } from 'react-bootstrap';
import logo from '../assets/sailor-logo.png';

interface ToastProps {
    message: string;
    duration?: number; // in milliseconds
    onClose?: () => void;
}

// TODO: still we are rendering one behind the other, instead we need the toast before one to exit and show the next one
const Toast: React.FC<ToastProps> = ({ message, duration = 3000, onClose }) => {
    const [show, setShow] = useState(true);

    useEffect(() => {
        if (show) {
            const timer = setTimeout(() => {
                setShow(false);
                if (onClose) onClose();
            }, duration);
            return () => clearTimeout(timer);
        }
    }, [show, duration, onClose]);

    return (
        <ToastContainer position="bottom-end" className="p-3" style={{ zIndex: 9999, width: 'inherit !important' }}>
            <BootstrapToast style={{ color: 'white' }} show={show} onClose={() => { setShow(false); if (onClose) onClose(); }} bg="secondary" delay={duration} autohide>
                <BootstrapToast.Header>
                    <img
                        style={{ width: 20, height: 20 }}
                        src={logo}
                        alt="sailor-logo"
                    />
                    <strong className="me-auto">sailor says...</strong>
                </BootstrapToast.Header>
                <BootstrapToast.Body>
                    {message}
                </BootstrapToast.Body>
            </BootstrapToast>
        </ToastContainer>
    );
};

export default Toast; 