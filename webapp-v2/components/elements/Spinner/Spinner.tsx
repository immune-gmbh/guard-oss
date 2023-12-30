const Spinner = (): JSX.Element => {
  return (
    <div
      className="fixed top-0 left-0 right-0 bottom-0 z-50 bg-purple-500 bg-opacity-90 flex justify-center items-center"
      role="status">
      <svg
        className="spinner h-2/5"
        viewBox="-5 -5 120 120"
        xmlns="http://www.w3.org/2000/svg"
        strokeLinecap="square">
        <defs>
          <linearGradient id="linear" x1="0%" y1="100%" x2="100%" y2="0%">
            <stop offset="0%" stopColor="#FF193C" stopOpacity="0.5" />
            <stop offset="100%" stopColor="#FF193C" />
          </linearGradient>
        </defs>

        <path
          stroke="url(#linear)"
          className="path"
          pathLength="100"
          strokeLinecap="round"
          d="M76 6.2l26.7 34.3a10 10 0 011.4 9.6L91.8 83.8a10 10 0 01-7.5 6.4L41 98.3a10 10 0 01-9.7-3.6L5.1 61.4a10 10 0 01-1.6-9.6L16 18a10 10 0 017.3-6.3L66 2.5a10 10 0 0110 3.7z"
        />
      </svg>
      <style>{`
        $pink: #ff193c;
        $beere: #410028;
        $stroke-length: 100;
        $dash: 20;
        $gap: $stroke-length - $dash;

        @keyframes colors {
          0% { stroke: $pink; }
          50% { stroke: $beere; }
          100% { stroke: $pink; }
        }

        @keyframes dash {
          0% {
            stroke-dashoffset: 100;
          }
          100% {
            stroke-dashoffset: 0;
          }
        }

        .path {
          margin: 0 auto;
          display: block;
          fill: none;
          stroke-width: 4;
          stroke-dasharray: 20 80;
          stroke-dashoffset: 0;
          stroke: url(#linear);
          transform-origin: 55px 55px;
          animation: dash 2s linear infinite;
        }
      `}</style>
    </div>
  );
};

export default Spinner;
