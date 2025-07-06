import { createRoot } from 'react-dom/client';
import Toast from '../components/Toast';

export function showToast(message: string, duration: number = 3) {
    const toastContainer = document.createElement('div');
    document.body.appendChild(toastContainer);
    const root = createRoot(toastContainer);

    const handleClose = () => {
        setTimeout(() => {
            root.unmount();
            document.body.removeChild(toastContainer);
        }, duration); // allow fade out
    };

    root.render(
        <Toast message={message} duration={duration} onClose={handleClose} />
    );
} 