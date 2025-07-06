import { configureStore } from '@reduxjs/toolkit';
import appsReducer from './appsSlice';
import backupReducer from './backupSlice';

const store = configureStore({
  reducer: {
    apps: appsReducer,
    backup: backupReducer,
  },
});

export type RootState = ReturnType<typeof store.getState>;
export type AppDispatch = typeof store.dispatch;
export default store; 