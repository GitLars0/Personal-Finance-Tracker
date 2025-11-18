import React from 'react';

export function TextInput({ label, value, onChange, placeholder, required }) {
  return (
    <div className="form-group">
      <label>{label}{required ? ' *' : ''}</label>
      <input type="text" value={value} onChange={e => onChange(e.target.value)} placeholder={placeholder} required={required} />
    </div>
  );
}

export function NumberInput({ label, value, onChange, placeholder, step = '0.01', required }) {
  return (
    <div className="form-group">
      <label>{label}{required ? ' *' : ''}</label>
      <input type="number" step={step} value={value} onChange={e => onChange(e.target.value)} placeholder={placeholder} required={required} />
    </div>
  );
}

export function DateInput({ label, value, onChange, required }) {
  return (
    <div className="form-group">
      <label>{label}{required ? ' *' : ''}</label>
      <input type="date" value={value} onChange={e => onChange(e.target.value)} required={required} />
    </div>
  );
}

export function SelectInput({ label, value, onChange, options = [], placeholder = 'Select', required }) {
  return (
    <div className="form-group">
      <label>{label}{required ? ' *' : ''}</label>
      <select value={value} onChange={e => onChange(e.target.value)} required={required}>
        <option value="">{placeholder}</option>
        {options.map(opt => (
          <option key={opt.value ?? opt.id ?? opt} value={opt.value ?? opt.id ?? opt}>
            {opt.label ?? opt.name ?? opt}
          </option>
        ))}
      </select>
    </div>
  );
}
