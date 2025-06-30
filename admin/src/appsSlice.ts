import { createSlice } from '@reduxjs/toolkit';

const appsSlice = createSlice({
    name: 'apps',
    initialState: {
        values: [] as string[],
    },
    reducers: {
        setApps: (state, action) => {
            state.values = action.payload;
        },
    },
});

export const { setApps } = appsSlice.actions;
export default appsSlice.reducer;