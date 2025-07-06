import { createSlice } from '@reduxjs/toolkit';

const backupSlice = createSlice({
    name: 'backup',
    initialState: {
        bucket: '',
    },
    reducers: {
        setBucket: (state, action) => {
            state.bucket = action.payload;
        },
    },
});

export const { setBucket } = backupSlice.actions;
export default backupSlice.reducer;